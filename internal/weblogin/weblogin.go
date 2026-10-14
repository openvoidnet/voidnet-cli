// Package weblogin implements browser-based sign-in for voidnet.
//
// Hand-written Void flow (mirrors the frozen login contract in
// login-contract/contract.yaml): PKCE loopback on 127.0.0.1, browser login at
// Void Accounts, authorization-code redemption, then a machine-token mint at
// the publisher API. Generated SDK clients are reused for the mint and the
// revocation; only the browser dance is hand code. Never edited by regen.
package weblogin

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/openvoidnet/voidnet-cli/internal/config"
	expsdk "github.com/openvoidnet/voidnet-cli/internal/sdk"
	"github.com/openvoidnet/voidnet-cli/internal/sdk/models/components"
	"github.com/openvoidnet/voidnet-cli/internal/sdk/models/operations"
	"github.com/openvoidnet/voidnet-cli/internal/sdk/models/publisher"
)

const (
	oauthClientID = "voidnet-cli"
	oauthScope    = "publisher_login"

	defaultAuthBase = "https://accounts.openvoidnet.com"
	defaultMgmtBase = "https://openvoidnet.com"

	devAuthBase = "http://localhost:3020"
	devMgmtBase = "http://localhost:3000"

	envAuthBase = "VOIDNET_AUTH_BASE"
	envMgmtBase = "VOIDNET_MGMT_BASE"

	callbackPath = "/callback"
	loginTimeout = 5 * time.Minute
	httpTimeout  = 30 * time.Second
	tokenKey     = "publisher-token"
)

// Endpoints resolves the two origins the flow touches.
type Endpoints struct {
	AuthBase string
	MgmtBase string
}

// ResolveEndpoints applies explicit bases > --env table > env vars > production default.
func ResolveEndpoints(envName, authBaseFlag, mgmtBaseFlag string) (Endpoints, error) {
	auth := authBaseFlag
	mgmt := mgmtBaseFlag
	if auth == "" && mgmt == "" {
		switch envName {
		case "development":
			auth, mgmt = devAuthBase, devMgmtBase
		case "production", "":
		default:
			return Endpoints{}, fmt.Errorf("invalid --env %q (production|development)", envName)
		}
	}
	if auth == "" {
		auth = os.Getenv(envAuthBase)
	}
	if auth == "" {
		auth = defaultAuthBase
	}
	if mgmt == "" {
		mgmt = os.Getenv(envMgmtBase)
	}
	if mgmt == "" {
		mgmt = defaultMgmtBase
	}
	return Endpoints{AuthBase: auth, MgmtBase: mgmt}, nil
}

// trace prints flow internals only under --debug. Default runs stay quiet:
// action-needed lines plus outcomes. Errors always print at call sites.
func trace(stderr io.Writer, debug bool, msg string) {
	if debug {
		fmt.Fprintln(stderr, "[weblogin] "+msg)
	}
}

// tokenResponse mirrors POST /api/auth/oauth/token for authorization_code.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

type callbackResult struct {
	code  string
	state string
	err   error
}

func newPKCEPair(stderr io.Writer, debug bool) (verifier, challenge, state string, err error) {
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return "", "", "", fmt.Errorf("pkce entropy failed: %w", err)
	}
	verifier = base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	stateRaw := make([]byte, 16)
	if _, err = rand.Read(stateRaw); err != nil {
		return "", "", "", fmt.Errorf("state entropy failed: %w", err)
	}
	state = base64.RawURLEncoding.EncodeToString(stateRaw)
	trace(stderr, debug, "pkce pair generated (S256)")
	return verifier, challenge, state, nil
}

func buildAuthorizeURL(authBase, redirectURI, challenge, state string) string {
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", oauthClientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", oauthScope)
	q.Set("state", state)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	return authBase + "/api/auth/oauth/authorize?" + q.Encode()
}

func openBrowser(target string, stderr io.Writer, debug bool) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", target)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	default:
		cmd = exec.Command("xdg-open", target)
	}
	fmt.Fprintln(stderr, "Opening browser for Void Accounts sign-in. If nothing opens, use this URL:")
	fmt.Fprintln(stderr, target)
	if err := cmd.Start(); err != nil {
		trace(stderr, debug, fmt.Sprintf("browser auto-open failed: %v", err))
		return
	}
	trace(stderr, debug, "browser auto-open attempted")
}

// serveCallback listens on 127.0.0.1:0 and returns the redirect URI plus a
// channel delivering exactly one provider result. Provider refusals
// (access_denied) surface as errors, never as exchanges.
func serveCallback(stderr io.Writer, debug bool) (redirectURI string, ch <-chan callbackResult, shutdown func()) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		failed := make(chan callbackResult, 1)
		failed <- callbackResult{err: fmt.Errorf("loopback listen failed: %w", err)}
		return "", failed, func() {}
	}
	redirectURI = "http://" + listener.Addr().String() + callbackPath
	results := make(chan callbackResult, 1)
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != callbackPath {
				http.NotFound(w, r)
				return
			}
			q := r.URL.Query()
			var out callbackResult
			if q.Get("error") != "" {
				out = callbackResult{err: fmt.Errorf("provider refused: %s", q.Get("error"))}
				fmt.Fprintln(w, "Sign-in refused. You can close this tab.")
			} else {
				out = callbackResult{code: q.Get("code"), state: q.Get("state")}
				fmt.Fprintln(w, "Signed in. You can close this tab and return to the terminal.")
			}
			select {
			case results <- out:
			default:
			}
		}),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		_ = server.Serve(listener)
	}()
	trace(stderr, debug, "loopback listener ready")
	return redirectURI, results, func() {
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutCtx)
	}
}

// redeemCode posts the authorization_code grant and returns the accounts token.
func redeemCode(ctx context.Context, authBase, code, redirectURI, verifier string, stderr io.Writer, debug bool) (string, error) {
	payload, err := json.Marshal(map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     oauthClientID,
		"code":          code,
		"redirect_uri":  redirectURI,
		"code_verifier": verifier,
	})
	if err != nil {
		return "", fmt.Errorf("token request encode failed: %w", err)
	}
	reqCtx, cancel := context.WithTimeout(ctx, httpTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, authBase+"/api/auth/oauth/token", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("token request build failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint http=%d body=%.200s", resp.StatusCode, string(raw))
	}
	var parsed tokenResponse
	if err := json.Unmarshal(raw, &parsed); err != nil || parsed.AccessToken == "" {
		return "", fmt.Errorf("token response undecodable")
	}
	trace(stderr, debug, "accounts token issued (1h)")
	return parsed.AccessToken, nil
}

// verifyMachineToken proves the token against the publisher API and returns
// the visible app count. Anything but a clean list is a failed verification.
func verifyMachineToken(ctx context.Context, mgmtBase, token string, stderr io.Writer, debug bool) (int, error) {
	client := expsdk.New(
		expsdk.WithSecurity(components.Security{PublisherToken: &token}),
	)
	callCtx, cancel := context.WithTimeout(ctx, httpTimeout)
	defer cancel()
	res, err := client.Publisher.ListApps(callCtx, operations.WithServerURL(mgmtBase))
	if err != nil {
		return 0, fmt.Errorf("token verification failed: %w", err)
	}
	if res == nil || res.Object == nil {
		return 0, fmt.Errorf("token verification returned empty apps")
	}
	n := len(res.Object.Apps)
	trace(stderr, debug, fmt.Sprintf("token verified against publisher API (apps=%d)", n))
	return n, nil
}

// rollbackMint revokes a just-minted token when post-mint verification fails,
// so a failed login never strands a live unknown token.
func rollbackMint(ctx context.Context, mgmtBase, token string, stderr io.Writer, debug bool) {
	client := expsdk.New(
		expsdk.WithSecurity(components.Security{PublisherToken: &token}),
	)
	callCtx, cancel := context.WithTimeout(ctx, httpTimeout)
	defer cancel()
	if _, err := client.Publisher.RevokeSelf(callCtx, operations.WithServerURL(mgmtBase)); err != nil {
		trace(stderr, debug, fmt.Sprintf("rollback revoke failed: %v", err))
		return
	}
	trace(stderr, debug, "unverified mint rolled back (revoked)")
}

// mintMachineToken calls the generated publisher exchange op. Storage and
// verification are separate steps so every entry path proves before it keeps.
func mintMachineToken(ctx context.Context, mgmtBase, accountsToken string, stderr io.Writer, debug bool) (token, id string, err error) {
	client := expsdk.New()
	callCtx, cancel := context.WithTimeout(ctx, httpTimeout)
	defer cancel()
	res, err := client.Publisher.ExchangeTokens(callCtx, publisher.ExchangeRequest{AccountsToken: accountsToken}, operations.WithServerURL(mgmtBase))
	if err != nil {
		return "", "", fmt.Errorf("machine token mint failed: %w", err)
	}
	if res == nil || res.ExchangeResult == nil || res.ExchangeResult.Token == "" {
		return "", "", fmt.Errorf("machine token mint returned empty token")
	}
	trace(stderr, debug, "machine token minted (30d, read+verify)")
	return res.ExchangeResult.Token, res.ExchangeResult.ID, nil
}

// storeMachineToken persists through the same store the generated CLI reads.
func storeMachineToken(token string, stderr io.Writer, debug bool) error {
	cfg := config.GetConfig()
	if cfg == nil {
		cfg = &config.Config{}
	}
	keychainStored := config.StoreSecret(tokenKey, token, &cfg.Security.PublisherToken) == nil
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("config save failed: %w", err)
	}
	if keychainStored {
		trace(stderr, debug, "machine token stored in OS keychain")
	}
	trace(stderr, debug, fmt.Sprintf("machine token saved (config: %s)", config.GetConfigPath()))
	return nil
}

// LoginWithToken verifies a directly supplied token before storing anything.
// Rejects garbage at the door instead of persisting a dead credential.
func LoginWithToken(ctx context.Context, out, stderr io.Writer, mgmtBase, token string, debug bool) error {
	if token == "" {
		return fmt.Errorf("empty token, nothing stored")
	}
	vctx, cancel := context.WithTimeout(ctx, httpTimeout)
	defer cancel()
	n, err := verifyMachineToken(vctx, mgmtBase, token, stderr, debug)
	if err != nil {
		return fmt.Errorf("supplied token rejected by publisher API, nothing stored: %w", err)
	}
	if err := storeMachineToken(token, stderr, debug); err != nil {
		return err
	}
	fmt.Fprintf(out, "Signed in. Token verified (apps=%d) and stored.\n", n)
	return nil
}

// revokeStoredToken revokes the currently stored machine token server-side,
// then wipes local credentials. Mirrors built-in logout wipe, plus the call.
func revokeStoredToken(ctx context.Context, mgmtBase string, stderr io.Writer, debug bool) error {
	cfg := config.GetConfig()
	if cfg == nil {
		cfg = &config.Config{}
	}
	stored := config.GetStoredSecret(tokenKey, cfg.Security.PublisherToken)
	if stored == "" {
		return fmt.Errorf("no stored machine token to revoke")
	}
	client := expsdk.New(
		expsdk.WithSecurity(components.Security{PublisherToken: &stored}),
	)
	callCtx, cancel := context.WithTimeout(ctx, httpTimeout)
	defer cancel()
	if _, err := client.Publisher.RevokeSelf(callCtx, operations.WithServerURL(mgmtBase)); err != nil {
		return fmt.Errorf("server revocation failed: %w", err)
	}
	fmt.Fprintln(stderr, "server revocation confirmed")
	_ = config.DeleteKeyringValue(tokenKey)
	cfg.Security.PublisherToken = ""
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("config save failed: %w", err)
	}
	fmt.Fprintln(stderr, "local credentials wiped")
	return nil
}

// Login runs the full browser dance and stores the machine token.
func Login(ctx context.Context, out, stderr io.Writer, authBase, mgmtBase string, debug bool) error {
	flowCtx, cancel := context.WithTimeout(ctx, loginTimeout)
	defer cancel()
	verifier, challenge, state, err := newPKCEPair(stderr, debug)
	if err != nil {
		return err
	}
	redirectURI, results, shutdown := serveCallback(stderr, debug)
	defer shutdown()
	if redirectURI == "" {
		failed := <-results
		return failed.err
	}
	openBrowser(buildAuthorizeURL(authBase, redirectURI, challenge, state), stderr, debug)
	fmt.Fprintln(stderr, "Waiting for browser approval...")
	var got callbackResult
	select {
	case got = <-results:
	case <-flowCtx.Done():
		return fmt.Errorf("browser callback timed out")
	}
	if got.err != nil {
		return got.err
	}
	if got.code == "" {
		return fmt.Errorf("callback carried no code")
	}
	if got.state != state {
		return fmt.Errorf("state mismatch, aborting before exchange")
	}
	trace(stderr, debug, "callback received, state verified")
	accountsToken, err := redeemCode(flowCtx, authBase, got.code, redirectURI, verifier, stderr, debug)
	if err != nil {
		return err
	}
	token, tokenID, err := mintMachineToken(flowCtx, mgmtBase, accountsToken, stderr, debug)
	if err != nil {
		return err
	}
	if _, verr := verifyMachineToken(flowCtx, mgmtBase, token, stderr, debug); verr != nil {
		rollbackMint(flowCtx, mgmtBase, token, stderr, debug)
		return fmt.Errorf("fresh mint failed verification, rolled back: %w", verr)
	}
	if err := storeMachineToken(token, stderr, debug); err != nil {
		rollbackMint(flowCtx, mgmtBase, token, stderr, debug)
		return err
	}
	fmt.Fprintf(out, "Signed in. Machine token %s verified and stored.\n", tokenID)
	return nil
}

// AuthHint returns the remedy line for API auth failures, shared by every
// rendering path so 401/403 never collapse into one message.
func AuthHint(statusCode int, rawBody string) string {
	switch {
	case strings.Contains(rawBody, "api_key_missing"):
		return "gateway needs a buyer credential (see --api-key-auth); publisher tokens only work on publisher ops"
	case strings.Contains(rawBody, "api_key_invalid"):
		return "buyer credential rejected (format: Bearer vnb-sk-* or Bearer JWT via --api-key-auth)"
	case strings.Contains(rawBody, "app_not_purchased"):
		return "app not purchased for this buyer (subscribe in Voidnet Marketplace, then retry)"
	case strings.Contains(rawBody, "publisher_error"):
		return "publisher app offline or errored upstream (owner: check app server_url and server logs)"
	case statusCode == 401:
		return "publisher token rejected (expired or revoked). Run: voidnet login"
	case statusCode == 403 && strings.Contains(rawBody, "insufficient_scope"):
		return "token lacks the required scope (login grants read+verify only; publish and keys:rotate stay in Voidnet Console)"
	case statusCode == 403 && strings.Contains(rawBody, "insufficient_binding"):
		return "token is not bound to this app (consent selection mismatch)"
	case statusCode == 401 || statusCode == 403:
		return "run 'voidnet configure' to set up or update your credentials"
	default:
		return ""
	}
}

// TranslateAPIError maps raw API failures to the three login truths so no
// caller can collapse them: absent credential, dead token, wrong scope.
// Anything unrecognized passes through untouched.
func TranslateAPIError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "api_key_missing"):
		return fmt.Errorf("gateway needs a buyer credential (see --api-key-auth); publisher tokens only work on publisher ops")
	case strings.Contains(msg, "api_key_invalid"):
		return fmt.Errorf("buyer credential rejected (format: Bearer vnb-sk-* or Bearer JWT via --api-key-auth)")
	case strings.Contains(msg, "app_not_purchased"):
		return fmt.Errorf("app not purchased for this buyer (subscribe in Voidnet Marketplace, then retry)")
	case strings.Contains(msg, "publisher_error"):
		return fmt.Errorf("publisher app offline or errored upstream (owner: check app server_url and server logs)")
	case containsStatus(msg, 401):
		return fmt.Errorf("publisher token rejected (expired or revoked). Run: voidnet login — %w", err)
	case containsStatus(msg, 403) && strings.Contains(msg, "insufficient_scope"):
		return fmt.Errorf("token lacks the required scope (login grants read+verify only; publish and keys:rotate stay in Voidnet Console) — %w", err)
	case containsStatus(msg, 403) && strings.Contains(msg, "insufficient_binding"):
		return fmt.Errorf("token is not bound to this app (consent selection mismatch) — %w", err)
	default:
		return err
	}
}

func containsStatus(msg string, status int) bool {
	for _, form := range []string{fmt.Sprintf("Status %d", status), fmt.Sprintf("status=%d", status), fmt.Sprintf("status %d", status)} {
		if strings.Contains(msg, form) {
			return true
		}
	}
	return false
}

// Logout revokes server-side, then wipes local state.
func Logout(ctx context.Context, out, stderr io.Writer, mgmtBase string, debug bool) error {
	if err := revokeStoredToken(ctx, mgmtBase, stderr, debug); err != nil {
		return err
	}
	fmt.Fprintln(out, "Signed out. Server token revoked, local credentials wiped.")
	return nil
}
