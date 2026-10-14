## voidnet mcp create-token

Request an OAuth 2.0 access token (MCP clients)

### Synopsis

Exchange credentials for a short-lived JWT access token using supported grant types: client_credentials, authorization_code, refresh_token. Built for OAuth-aware MCP clients (Claude Desktop, Cursor, VS Code); default scope is mcp:tools. The token is valid for a configurable TTL (default 1 hour) and can be used as a `Bearer` credential anywhere an API key is accepted. The `audience` is always the gateway issuer URL — no client-controlled audience is accepted.

```
voidnet mcp create-token [flags]
```

### Examples

```
  voidnet mcp create-token --client-id <value> --grant-type client_credentials
```

### Options

```
      --body string            Request body as JSON (alternative to individual flags). Can also be provided via stdin; @path reads a file, @- reads stdin to EOF. Use --schema to print the exact JSON Schema.
      --client-id string       Your API key (vnb-sk-* / vai-sk-* / vnp-sk-*) [required]
      --client-secret string   Must match client_id if provided (API key is its own secret)
      --code string            Authorization code (required for authorization_code grant)
      --code-verifier string   PKCE code verifier (required for authorization_code grant)
  -g, --grant-type string      OAuth grant type (options: client_credentials, authorization_code, refresh_token) [required]
  -h, --help                   help for create-token
      --redirect-uri string    Redirect URI (required for authorization_code grant)
      --refresh-token string   Refresh token (required for refresh_token grant)
      --schema                 Print the exact JSON Schema of the request body and exit
  -s, --scope string           Requested scope (mcp:* for tools, llm:* for models) (default "mcp:tools")
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

* `voidnet mcp create-token --usage` — this command's flags, defaults and env vars as machine-readable KDL
* `voidnet mcp create-token --schema` — the exact JSON Schema of the request body (all `$ref`s bundled)
* `voidnet mcp create-token --dry-run` — preview the request without OS-keychain access or a network call (human preview on stderr)
* `--dry-run --output-format json` (or a caller-explicit `--jq`) writes one preview object per request as NDJSON on stdout; jq is not applied to previews
* `--output-format json` or `--jq <expr>` for machine-readable live output; in agent mode errors are a JSON envelope on stderr

Exit codes: 0 ok · 1 runtime · 2 usage · 3 authentication/authorization
