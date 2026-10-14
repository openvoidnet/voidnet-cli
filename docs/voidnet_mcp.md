## voidnet mcp

MCP tools over JSON-RPC

### Synopsis

MCP tools over JSON-RPC

```
voidnet mcp [flags]
```

### Options

```
  -h, --help   help for mcp
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

* [voidnet](voidnet.md)	 - voidnet: Sign in once, manage publisher apps, call buyer APIs
* [voidnet mcp call-tool](voidnet_mcp_call-tool.md)	 - Call an MCP tool (stateless 2026-07-28)
* [voidnet mcp create-token](voidnet_mcp_create-token.md)	 - Request an OAuth 2.0 access token (MCP clients)
* [voidnet mcp get-authorization-server-metadata](voidnet_mcp_get-authorization-server-metadata.md)	 - OAuth authorization-server metadata (RFC 8414)
* [voidnet mcp get-authorize-page](voidnet_mcp_get-authorize-page.md)	 - OAuth login+consent page (Authorization Code + PKCE)
* [voidnet mcp get-jwks](voidnet_mcp_get-jwks.md)	 - JSON Web Key Set (JWKS)
* [voidnet mcp get-protected-resource-metadata](voidnet_mcp_get-protected-resource-metadata.md)	 - OAuth protected-resource metadata (RFC 9728)
* [voidnet mcp query](voidnet_mcp_query.md)	 - Call an MCP tool (stateless 2026-07-28)
* [voidnet mcp submit-authorize](voidnet_mcp_submit-authorize.md)	 - Submit login+consent, issue authorization code

Exit codes: 0 ok · 1 runtime · 2 usage · 3 authentication/authorization
