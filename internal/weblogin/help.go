// Curated front-page help (git-style): main commands up front, everything
// behind -a. Secondary commands are Hidden at runtime (still runnable),
// so both `help` and `--help` stay scannable. Never edited by regen.
package weblogin

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// mainCommands is the front-page set. Everything else top-level hides.
var mainCommands = map[string]bool{
	"login":     true,
	"logout":    true,
	"publisher": true,
	"mcp":       true,
	"llm":       true,
	"health":    true,
	"whoami":    true,
	"profile":   true,
	"help":      true,
}

// HideSecondaryCommands marks non-main top-level commands Hidden. Hidden
// commands stay fully runnable and appear under `help -a`.
func HideSecondaryCommands(root *cobra.Command) {
	for _, c := range root.Commands() {
		if !mainCommands[c.Name()] {
			c.Hidden = true
		}
	}
}

type helpSection struct {
	title string
	cmds  []string
}

var helpSections = []helpSection{
	{"sign in", []string{"login", "logout"}},
	{"manage", []string{"publisher"}},
	{"consume", []string{"mcp", "llm"}},
	{"check", []string{"health", "whoami", "profile"}},
}

// HelpCommand replaces the default help command. Bare `help` shows the
// curated front page, `help -a` lists every command, `help <topic>` shows
// that command's help (including hidden topics).
func HelpCommand(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "help [command]",
		Short: "Help about commands (use -a for all)",
		Long:  "Show the curated command list. Use help -a for every command, or help <command> for details.",
		Args:  cobra.ArbitraryArgs,
	}
	var showAll bool
	cmd.Flags().BoolVarP(&showAll, "all", "a", false, "List every command, including utilities")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		if len(args) > 0 {
			target, _, err := root.Find(args)
			if err != nil || target == nil {
				return fmt.Errorf("unknown help topic %q (try: help -a)", strings.Join(args, " "))
			}
			return target.Help()
		}
		if showAll {
			return fullHelp(out, root)
		}
		return frontHelp(out, root)
	}
	return cmd
}

func frontHelp(out io.Writer, root *cobra.Command) error {
	byName := map[string]*cobra.Command{}
	for _, c := range root.Commands() {
		byName[c.Name()] = c
	}
	fmt.Fprintln(out, "Sign in once, manage publisher apps, call buyer APIs.")
	fmt.Fprintln(out, "")
	for _, s := range helpSections {
		fmt.Fprintf(out, "%s:\n", s.title)
		for _, name := range s.cmds {
			c, ok := byName[name]
			if !ok {
				continue
			}
			fmt.Fprintf(out, "  %-10s %s\n", name, c.Short)
		}
		fmt.Fprintln(out, "")
	}
	fmt.Fprintln(out, "'help -a' lists every command. 'help <command>' shows details.")
	return nil
}

func fullHelp(out io.Writer, root *cobra.Command) error {
	names := []string{}
	seen := map[string]bool{}
	var walk func(c *cobra.Command, prefix string)
	walk = func(c *cobra.Command, prefix string) {
		for _, sub := range c.Commands() {
			path := prefix + sub.Name()
			if seen[path] {
				continue
			}
			seen[path] = true
			names = append(names, path)
			walk(sub, path+" ")
		}
	}
	walk(root, "")
	sort.Strings(names)
	fmt.Fprintln(out, "Every command (main + utilities):")
	fmt.Fprintln(out, "")
	for _, n := range names {
		fmt.Fprintf(out, "  %s\n", n)
	}
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "'help <command>' shows details.")
	return nil
}
