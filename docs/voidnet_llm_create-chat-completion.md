## voidnet llm create-chat-completion

Chat with an AI model (OpenAI compatible, text + multimodal)

### Synopsis

Proxy an OpenAI-compatible chat request to the target LLM app's publisher server. Body: `{"model":"<published-model-id>","messages":[{"role":"user","content":"..."}]}` — `model` must match the app's published model exactly. Text plus multimodal content parts (`image_url`, `input_audio`, `video_url`, `file`) pass through verbatim to the publisher. `stream:false` and `stream:true` (SSE) both supported; the gateway injects `stream_options.include_usage:true` when missing. Streams without a final usage chunk fail loudly instead of going unbilled. One model per app, token-metered (`usage.total_tokens`), wallet-billed per call.

```
voidnet llm create-chat-completion [flags]
```

### Examples

```
  voidnet llm create-chat-completion --username Chadd.Jacobi85 --appname <value> --body '{"key": "value"}'
```

### Options

```
      --api-key-auth string   Security credential
  -a, --appname string        App name, unique per publisher [required]
      --body string           Request body as JSON (alternative to individual flags). Can also be provided via stdin; @path reads a file, @- reads stdin to EOF. Use --schema to print the exact JSON Schema.
  -h, --help                  help for create-chat-completion
  -r, --request-body string   OpenAI chat body {model, messages} — proxied verbatim [required]
      --schema                Print the exact JSON Schema of the request body and exit
  -u, --username string       Publisher username (the developer's public identifier) [required]
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

* [voidnet llm](voidnet_llm.md)	 - Chat + models (OpenAI-compatible)

### Machine interface

* `voidnet llm create-chat-completion --usage` — this command's flags, defaults and env vars as machine-readable KDL
* `voidnet llm create-chat-completion --schema` — the exact JSON Schema of the request body (all `$ref`s bundled)
* `voidnet llm create-chat-completion --dry-run` — preview the request without OS-keychain access or a network call (human preview on stderr)
* `--dry-run --output-format json` (or a caller-explicit `--jq`) writes one preview object per request as NDJSON on stdout; jq is not applied to previews
* `--output-format json` or `--jq <expr>` for machine-readable live output; in agent mode errors are a JSON envelope on stderr

Exit codes: 0 ok · 1 runtime · 2 usage · 3 authentication/authorization
