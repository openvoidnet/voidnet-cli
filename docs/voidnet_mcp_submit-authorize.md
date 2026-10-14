## voidnet mcp submit-authorize

Submit login+consent, issue authorization code

### Synopsis

Verify the user's email+password (bcrypt), then issue a single-use authorization code (5-minute TTL) and redirect (302) to redirect_uri with ?code=...&state=.... Validation failures re-render the HTML login form with an error message instead of returning JSON.

```
voidnet mcp submit-authorize [flags]
```

### Examples

```
  voidnet mcp submit-authorize --client-id <value> --code-challenge <value> --code-challenge-method <value> --email <value> --password <value>
```

### Options

```
      --body string                    Request body as JSON (alternative to individual flags). Can also be provided via stdin; @path reads a file, @- reads stdin to EOF.
      --client-id string               Buyer API key carried from the GET request [required]
      --code-challenge string          PKCE challenge carried from the GET request [required]
      --code-challenge-method string   Must be 'S256' [required]
  -e, --email string                   Console account email [required]
  -h, --help                           help for submit-authorize
  -p, --password string                Console account password [required]
  -r, --redirect-uri string            Exact-match http/https callback URL [required]
      --scope string                   Requested scope
      --state string                   Opaque client state, round-tripped to the callback
```

### Options inherited from parent commands

```
      --agent-mode               Enable structured errors and default TOON output for AI coding agents. Automatically enabled when a known agent environment is detected (CLAUDECODE, CURSOR_AGENT, etc.). Use --agent-mode=false to disable.
      --color string             Control colored output: auto (color when output is a TTY), always, or never. Respects NO_COLOR and FORCE_COLOR env vars. (default "auto")
  -d, --debug                    Log request and response diagnostics to stderr
      --dry-run                  Preview API requests without sending them (no network, no OS keychain). Human preview on stderr; with -o json or --jq, one JSON object per request on stdout. Local mutation commands (auth login, auth logout and configure) make no request: they skip prompts and writes and report a no-op (stderr, or one JSON object on stdout in the machine form)
  -H, --header stringArray       Set a custom HTTP request header (format: "Key: Value"). Can be specified multiple times.
      --include-headers          Include HTTP response headers in the output
      --interactive              Prompt for missing inputs and open guided configure/auth forms (forms fall back to line prompts on stdin off-TTY) (default true)
  -q, --jq string                Filter and transform output using a jq expression (e.g., '.name', '.items[] | .id')
      --no-interactive           Disable all interactive features (auto-prompting, explorer auto-launch, TUI forms)
  -o, --output-format string     Specify the output format. Options: pretty, json, yaml, table, toon. (default "pretty")
      --publisher-token string   HTTP Bearer
      --raw-output               Write --jq string results as raw text instead of JSON strings (like jq -r); non-string results stay JSON
      --server string            Select a server by index (for indexed servers) or name (for named servers)
      --server-url string        Override the default server URL
      --timeout string           HTTP request timeout (e.g., 30s, 5m, 100ms)
      --usage                    Print the CLI Usage schema in KDL format
```

### SEE ALSO

* [voidnet mcp](voidnet_mcp.md)	 - MCP tools over JSON-RPC

### Machine interface

* `voidnet mcp submit-authorize --usage` — this command's flags, defaults and env vars as machine-readable KDL
* `voidnet mcp submit-authorize --dry-run` — preview the request without OS-keychain access or a network call (human preview on stderr)
* `--dry-run --output-format json` (or a caller-explicit `--jq`) writes one preview object per request as NDJSON on stdout; jq is not applied to previews
* `--output-format json` or `--jq <expr>` for machine-readable live output; in agent mode errors are a JSON envelope on stderr

Exit codes: 0 ok · 1 runtime · 2 usage · 3 authentication/authorization
