// Command wiring for the hand-written browser login.
//
// Lives outside generated files so regen never touches it. The two-line seam
// in internal/cli/root.go (import + InitWebloginRoot call) is applied by
// scripts/wire-weblogin.patch after every generation run.
package weblogin

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/openvoidnet/voidnet-cli/internal/client"
	"golang.org/x/term"
)

// dryRunNoop honors the framework contract: no network, no keychain, no
// writes. Human form on stderr, one JSON object on stdout in machine form.
func dryRunNoop(cmd *cobra.Command, action, detail string) error {
	if client.IsJSONDryRun(cmd) {
		fmt.Fprintf(cmd.OutOrStdout(), "{\"dry_run\":true,\"action\":%q,\"network\":\"none\",\"writes\":\"none\"}\n", action)
		return nil
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Dry run: %s not executed (no network, no keychain, no writes)\n", action)
	fmt.Fprintf(cmd.ErrOrStderr(), "Plan: %s\n", detail)
	return nil
}

// isMachineToken enforces the vnp-pub-* shape before anything is stored.
func isMachineToken(tok string) bool {
	return strings.HasPrefix(tok, "vnp-pub-") && len(tok) > len("vnp-pub-")
}

// resolveUsername returns the flag value, or prompts once on interactive
// terminals. Empty stays empty: the server 404s honestly on a wrong name,
// and `voidnet profile` shows what is stored.
func resolveUsername(cmd *cobra.Command) (string, error) {
	if u, _ := cmd.Flags().GetString("username"); u != "" {
		return strings.TrimSpace(u), nil
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", nil
	}
	fmt.Fprint(cmd.ErrOrStderr(), "Publisher username (for buyer URLs, Enter to skip): ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(line), nil
}

// InitWebloginRoot registers top-level login/logout, mirroring the hand CLI
// surface and the speakeasy auth login UX.
func InitWebloginRoot(parent *cobra.Command) error {
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Sign in with Void Accounts in the browser",
		Long: `Opens the browser for Void Accounts sign-in, approves command-line access on the consent screen, and stores the minted machine token where every command reads it. The mint is verified against the publisher API before storing; a failed verification rolls the mint back.

login --token takes a console-issued vnp-pub-* token instead (headless/CI). It is verified against the publisher API before anything is stored; garbage is rejected at the door.

Flags override environment, environment overrides production defaults:
  --env production|development (sets both origins; explicit base flags win)
  --auth-base / VOIDNET_AUTH_BASE (default https://accounts.openvoidnet.com)
  --mgmt-base / VOIDNET_MGMT_BASE (default https://openvoidnet.com)`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			envName, _ := cmd.Flags().GetString("env")
			authBase, _ := cmd.Flags().GetString("auth-base")
			mgmtBase, _ := cmd.Flags().GetString("mgmt-base")
			ep, err := ResolveEndpoints(envName, authBase, mgmtBase)
			if err != nil {
				return err
			}
			if client.IsDryRun(cmd) {
				return dryRunNoop(cmd, "login", "open browser for Void Accounts sign-in, approve consent, redeem code, mint + verify + store machine token")
			}
			username, err := resolveUsername(cmd)
			if err != nil {
				return err
			}
			if tok, _ := cmd.Flags().GetString("token"); tok != "" {
				if !isMachineToken(tok) {
					return fmt.Errorf("--token must start with vnp-pub-, nothing stored")
				}
				return LoginWithToken(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), ep.MgmtBase, tok, username, client.IsDebug(cmd))
			}
			return Login(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), ep.AuthBase, ep.MgmtBase, username, client.IsDebug(cmd))
		},
	}
	loginCmd.Flags().String("auth-base", "", "Void Accounts origin (overrides VOIDNET_AUTH_BASE)")
	loginCmd.Flags().String("token", "", "Store a console-issued vnp-pub-* token directly (verified before storing; headless/CI)")
	loginCmd.Flags().String("mgmt-base", "", "Publisher API origin (overrides VOIDNET_MGMT_BASE)")
	loginCmd.Flags().String("env", "", "Environment shortcut: production|development (sets both origins; explicit base flags win)")
	loginCmd.Flags().String("username", "", "Publisher username to store (skips the prompt)")

	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Revoke the machine token and wipe local credentials",
		Long: `Revokes the stored machine token server-side, then wipes local credentials (keychain + config).

Flags override environment, environment overrides production defaults:
  --mgmt-base / VOIDNET_MGMT_BASE (default https://openvoidnet.com)`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			envName, _ := cmd.Flags().GetString("env")
			mgmtBase, _ := cmd.Flags().GetString("mgmt-base")
			ep, err := ResolveEndpoints(envName, "", mgmtBase)
			if err != nil {
				return err
			}
			if client.IsDryRun(cmd) {
				return dryRunNoop(cmd, "logout", "revoke machine token server-side, wipe keychain + config")
			}
			return Logout(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), ep.MgmtBase, client.IsDebug(cmd))
		},
	}
	logoutCmd.Flags().String("mgmt-base", "", "Publisher API origin (overrides VOIDNET_MGMT_BASE)")
	logoutCmd.Flags().String("env", "", "Environment shortcut: production|development (explicit --mgmt-base wins)")

	profileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Show the stored login profile",
		Long:  `Shows what login learned and stored: username, granted scopes, app binding, token expiry, and verification time. Nothing secret is printed. Absent profile means never logged in on this machine.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prof, err := LoadProfile()
			if err != nil {
				return fmt.Errorf("no stored profile — run voidnet login")
			}
			out := cmd.OutOrStdout()
			user := prof.Username
			if user == "" {
				user = "(not set — pass --username at login)"
			}
			fmt.Fprintf(out, "Username: %s\n", user)
			if len(prof.Scopes) > 0 {
				fmt.Fprintf(out, "Scopes: %s\n", strings.Join(prof.Scopes, ", "))
			} else {
				fmt.Fprintln(out, "Scopes: unknown (token supplied directly)")
			}
			fmt.Fprintf(out, "Apps: %s\n", prof.BindingLine())
			fmt.Fprintf(out, "Token: %s\n", prof.ExpiryStatus(time.Now()))
			if prof.VerifiedAt != "" {
				fmt.Fprintf(out, "Verified: %s\n", prof.VerifiedAt)
			}
			return nil
		},
	}

	for _, sub := range []*cobra.Command{loginCmd, logoutCmd, profileCmd} {
		exists := false
		for _, current := range parent.Commands() {
			if current.Name() == sub.Name() {
				exists = true
				break
			}
		}
		if !exists {
			parent.AddCommand(sub)
		}
	}
	// Curated front page: secondary commands hide (still runnable),
	// and `help` renders groups with -a for the full tree. Runs eagerly
	// for groups registered so far; FinalizeHelp covers the rest at the
	// end of root init (see wire-weblogin.patch).
	HideSecondaryCommands(parent)
	parent.SetHelpCommand(HelpCommand(parent))
	return nil
}

// FinalizeHelp re-applies hiding after all groups registered. Idempotent.
func FinalizeHelp(root *cobra.Command) {
	for _, c := range root.Commands() {
		if c.Name() == "help" && c.Long != "Show the curated command list. Use help -a for every command, or help <command> for details." {
			root.RemoveCommand(c)
		}
	}
	HideSecondaryCommands(root)
	root.SetHelpCommand(HelpCommand(root))
}
