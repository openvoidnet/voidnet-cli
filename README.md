# voidnet CLI

Access Voidnet API via CLI — `api.openvoidnet.com`.

Sign in once, manage publisher apps, call buyer APIs. Single binary, no runtime dependencies.

## Install

```sh
curl -fsSL https://openvoidnet.com/install.sh | sh
```

Or download a versioned binary from [Releases](https://github.com/openvoidnet/voidnet-cli/releases).

## First run

```sh
voidnet login
# opens the browser for Void Accounts sign-in; approve, return to terminal

voidnet whoami
voidnet publisher list-apps -o json
voidnet mcp call-tool --username <you> --appname <app> \
  --api-key-auth "Bearer <buyer-key>" \
  --body '{"jsonrpc":"2.0","id":"1","method":"tools/list"}'

voidnet logout   # revokes server-side, wipes local credentials
```

Headless/CI: `voidnet login --token vnp-pub-...` (verified before storing).
Dry run anything: append `--dry-run` (no network, no writes).
Full list: `voidnet help` (essentials), `voidnet help -a` (everything).

## Layout

* `cmd/voidnet/` — binary entrypoint.
* `internal/cli/` — generated Cobra commands (buyer + publisher operations).
* `internal/sdk/` — generated Go SDK the commands call.
* `internal/weblogin/` — hand-written browser login/logout (PKCE loopback, verify-before-store).
* `internal/output/` — pretty/json/yaml/table/toon rendering, error taxonomy.
* `login-contract/` — frozen login interface (see `scripts/check-login-contract.sh`).
* `overlays/`, `.speakeasy/` — spec merge + generation config. Regenerate with `scripts/generate.sh` (needs a Speakeasy account; output is deterministic).
* `scripts/check-cli-matrix.sh` — behavior matrix (needs local Voidnet stack + a test token).

## License

Apache-2.0. See `LICENSE`.

<!-- Start Summary [summary] -->
## Summary

voidnet: Sign in once, manage publisher apps, call buyer APIs.
<!-- End Summary [summary] -->

<!-- Start Table of Contents [toc] -->
## Table of Contents
<!-- $toc-max-depth=2 -->
* [voidnet CLI](#voidnet-cli)
  * [Install](#install)
  * [First run](#first-run)
  * [Layout](#layout)
  * [License](#license)
  * [CLI Installation](#cli-installation)
  * [Shell Completion](#shell-completion)
  * [CLI Example Usage](#cli-example-usage)
  * [For AI agents](#for-ai-agents)
  * [Authentication](#authentication)
  * [Commands](#commands)
  * [Request Body Input](#request-body-input)
  * [Output Formats](#output-formats)
  * [Error Handling](#error-handling)
  * [Diagnostics](#diagnostics)

<!-- End Table of Contents [toc] -->

<!-- Start CLI Installation [installation] -->
## CLI Installation

### Quick Install (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/openvoidnet/voidnet-cli/main/scripts/install.sh | bash
```

### Quick Install (Windows PowerShell)

```powershell
iwr -useb https://raw.githubusercontent.com/openvoidnet/voidnet-cli/main/scripts/install.ps1 | iex
```

### Go Install

Alternatively, install directly via Go:

```bash
go install github.com/openvoidnet/voidnet-cli/cmd/voidnet@latest
```

### Manual Download

Download pre-built binaries for your platform from the [releases page](https://github.com/openvoidnet/voidnet-cli/releases).
<!-- End CLI Installation [installation] -->

<!-- Start Shell Completion [completion] -->
## Shell Completion

Shell completions are available for Bash, Zsh, Fish, and PowerShell.

### Bash

```bash
# Add to ~/.bashrc:
source <(voidnet completion bash)

# Or install permanently:
voidnet completion bash > /etc/bash_completion.d/voidnet
```

### Zsh

```zsh
# Add to ~/.zshrc:
source <(voidnet completion zsh)

# Or install permanently:
voidnet completion zsh > "${fpath[1]}/_voidnet"
```

### Fish

```fish
voidnet completion fish | source

# Or install permanently:
voidnet completion fish > ~/.config/fish/completions/voidnet.fish
```

### PowerShell

```powershell
voidnet completion powershell | Out-String | Invoke-Expression
```
<!-- End Shell Completion [completion] -->

<!-- Start CLI Example Usage [usage] -->
## CLI Example Usage

### Example

```bash
voidnet mcp call-tool --api-key-auth test_api_key --mcp-protocol-version 2026-07-28 --mcp-method tools/call --mcp-name tools/call --appname '<value>' --id 1 --jsonrpc 2.0 --method tools/call --username Catherine_Lindgren28

```
<!-- End CLI Example Usage [usage] -->

<!-- Start For AI agents [agents] -->
## For AI agents

This CLI is built to be driven by AI coding agents as well as people: everything an agent needs is discoverable from the binary itself, and every command can be validated without credentials. Work down this ladder:

| Run | You get |
|-----|---------|
| `voidnet --help`, `voidnet publisher list-apps --help` | Commands by category, runnable examples, flags |
| `voidnet --usage`, `voidnet publisher list-apps --usage` | The command surface as machine-readable [KDL](https://kdl.dev): commands, aliases, flags, defaults, env vars, config keys |
| `voidnet mcp create-token --schema` | The exact JSON Schema of the command's request body (all `$ref`s bundled) — build a valid `--body` from it |
| `voidnet publisher list-apps --dry-run` | The exact HTTP request (method, URL, headers, body), with no credentials or network call |
| `voidnet publisher list-apps --output-format json` (or `--jq`) | Machine-readable output |

### Discover the command surface

```bash
# Every command, flag, default, env var and config key, as KDL
voidnet --usage

# One command's subtree only
voidnet publisher list-apps --usage
```

### Read the exact request schema

`--schema` is available on every command that accepts a request body (`--body`, stdin, or a whole-body flag where the command has one), including intent commands. It prints the JSON Schema the request is validated against and exits without calling the API.

```bash
# JSON Schema (draft 2020-12) of the request body, with every $ref bundled under $defs
voidnet mcp create-token --schema
```

### Probe before you spend

Start quota-spending commands with `--dry-run`. It validates inputs, resolves the request, redacts secrets and binary payloads, makes no network call, and exits 0. It never reads the OS keychain; credentials supplied by flag, environment, or config file are included only as `[REDACTED]`.

```bash
# Human preview: the [DRY-RUN] block is on stderr and stdout is empty
voidnet publisher list-apps --dry-run

# Machine preview: compact JSON on stdout and silent stderr
voidnet publisher list-apps --dry-run --output-format json
```

The machine form writes one object per would-be request, one per line (NDJSON for multi-request commands), with exactly this shape:

```json
{"dry_run":true,"request":{"method":"POST","url":"https://…","headers":{"Accept":["application/json"],…},"body":<JSON value | string | null>}}
```

`body` is a parsed JSON value when the body is JSON, a string for text, `"<bytes:N>"` for binary data, and `null` when absent. An explicit caller `--jq` also selects this JSON preview protocol, but the filter is not applied to preview objects. Command-declared jq presets do not select or filter the preview.

Local mutation commands make no request under `--dry-run`: instead of a preview they emit one `{"dry_run":true,"local":true,"command":"…","message":"…"}` object. `select(.request)` keeps only would-be requests; `select(.local)` keeps the local no-ops.

### Machine-readable output

```bash
# JSON on stdout
voidnet publisher list-apps --output-format json

# Filter or reshape with a jq expression (always emits JSON, overrides --output-format)
voidnet publisher list-apps --jq '.'

# Print jq string results as plain text instead of JSON strings (like jq -r)
voidnet publisher list-apps --jq '.' --raw-output
```

`--output-format toon` emits [TOON](https://github.com/toon-format/spec), a compact line-oriented format that uses fewer tokens than JSON; it is the default in agent mode.

### Interactive mode
Required-input prompts and guided `configure` / `auth login` forms are enabled by default. Required-input prompts require an interactive terminal; off-TTY forms read line input from stdin. Use `--no-interactive` to force flag-only execution.

```bash
# Prompt for missing command inputs
voidnet publisher list-apps --interactive

# Open the guided configuration form
voidnet configure --interactive

# Explicitly launch the terminal command explorer
voidnet explore
```

### Agent mode and structured errors
Agent mode turns on automatically when a known agent environment is detected (`CLAUDECODE`, `CURSOR_AGENT`, `CODEX`, `AIDER`, `CLINE`, `WINDSURF_AGENT`, `GITHUB_COPILOT`, `AMAZON_Q`, `GEMINI_CODE_ASSIST`, `SRC_CODY`) or with `--agent-mode` (`--agent-mode=false` disables detection).
In agent mode interactive prompts never launch, output defaults to TOON, and every failure — API errors and CLI usage errors alike — is one JSON envelope on stderr:
Outside agent mode, explicit JSON and `--jq` preserve the compatibility envelope without classification; enable agent mode to request the classified contract.

```json
{
  "error": "...",
  "error_type": "validation_error",
  "error_reason": "CLI_VALIDATION",
  "exit_code": 2,
  "message": "human-readable message",
  "hints": ["what to try next"]
}
```

`error_type` is one of `authentication_error`, `authorization_error`, `not_found`, `validation_error`, `rate_limit_error`, `server_error`, `api_error`, `connection_error`, `protocol_error`, `runtime_error`, `unsupported_error`, `async_failed`, `async_timeout`, `async_unknown_state`. Classification derives from the HTTP status and transport evidence; `error_reason` is absent for API errors. Status-less local failures may use `CLI_VALIDATION`, `CLI_CONNECTION`, `CLI_PROTOCOL`, `CLI_RUNTIME`, `CLI_UNAVAILABLE`, `CLI_AUTHENTICATION`, or the async polling reasons `CLI_ASYNC_FAILED`, `CLI_ASYNC_TIMEOUT`, and `CLI_ASYNC_UNKNOWN_STATE`. `hints` preserves server guidance first, adds the most specific local taxonomy guidance, then typed CLI and command-specific guidance, removing exact duplicates. `exit_code` is always the code for the final `error_type` shown in the envelope: 1 runtime, 2 usage, or 3 authentication/authorization.
<!-- End For AI agents [agents] -->

<!-- Start Authentication [security] -->
## Authentication

Authentication credentials can be configured in four ways (in order of priority):

### 1. Command-line flags

Pass credentials directly as flags to any command:

```bash
voidnet --publisher-token "$VOIDNET_PUBLISHER_TOKEN" publisher list-apps
```

### 2. Environment variables

Set credentials via environment variables:

| Variable | Description |
|----------|-------------|
| `VOIDNET_PUBLISHER_TOKEN` | HTTP Bearer |

### 3. OS Keychain (recommended for workstations)

Credentials are stored securely in your operating system's keychain when you run:

```bash
voidnet configure
```

Secret credentials (tokens, API keys, passwords) are automatically stored in:
- **macOS**: Keychain
- **Linux**: GNOME Keyring / KWallet (via D-Bus Secret Service)
- **Windows**: Windows Credential Locker

If no keychain is available (e.g., in CI environments), credentials fall back to the config file.

### 4. Configuration file

Run the interactive `configure` command to store non-secret settings:

```bash
voidnet configure
```

Configuration is stored in `~/.config/voidnet/config.yaml`.
<!-- End Authentication [security] -->

<!-- Start Commands [operations] -->
## Commands

<details open>
<summary>Available commands</summary>

* [`mcp`](docs/voidnet_mcp.md) - MCP tools over JSON-RPC
  * [`call-tool`](docs/voidnet_mcp_call-tool.md) - Call an MCP tool (stateless 2026-07-28)
  * [`create-token`](docs/voidnet_mcp_create-token.md) - Request an OAuth 2.0 access token (MCP clients)
  * [`get-authorization-server-metadata`](docs/voidnet_mcp_get-authorization-server-metadata.md) - OAuth authorization-server metadata (RFC 8414)
  * [`get-authorize-page`](docs/voidnet_mcp_get-authorize-page.md) - OAuth login+consent page (Authorization Code + PKCE)
  * [`get-jwks`](docs/voidnet_mcp_get-jwks.md) - JSON Web Key Set (JWKS)
  * [`get-protected-resource-metadata`](docs/voidnet_mcp_get-protected-resource-metadata.md) - OAuth protected-resource metadata (RFC 9728)
  * [`query`](docs/voidnet_mcp_query.md) - Call an MCP tool (stateless 2026-07-28)
  * [`submit-authorize`](docs/voidnet_mcp_submit-authorize.md) - Submit login+consent, issue authorization code
* [`publisher`](docs/voidnet_publisher.md) - Operations for publisher
  * [`exchange-tokens`](docs/voidnet_publisher_exchange-tokens.md) - Exchange accounts token for machine token
  * [`list-apps`](docs/voidnet_publisher_list-apps.md) - List publisher apps
  * [`list-keys`](docs/voidnet_publisher_list-keys.md) - List key fingerprints
  * [`publish-app`](docs/voidnet_publisher_publish-app.md) - Publish app
  * [`read-caller-profile`](docs/voidnet_publisher_read-caller-profile.md) - Caller identity and publisher state
  * [`read-metering`](docs/voidnet_publisher_read-metering.md) - Read metering rows
  * [`refresh-snapshot`](docs/voidnet_publisher_refresh-snapshot.md) - Refresh capability snapshot
  * [`revoke-self`](docs/voidnet_publisher_revoke-self.md) - Revoke calling token
  * [`rotate-key`](docs/voidnet_publisher_rotate-key.md) - Rotate app key
  * [`verify-server`](docs/voidnet_publisher_verify-server.md) - Verify publisher server
* [`health`](docs/voidnet_health.md) - Service health
  * [`check`](docs/voidnet_health_check.md) - Health check
  * [`check-liveness`](docs/voidnet_health_check-liveness.md) - Liveness probe
* [`llm`](docs/voidnet_llm.md) - Chat + models (OpenAI-compatible)
  * [`create-chat-completion`](docs/voidnet_llm_create-chat-completion.md) - Chat with an AI model (OpenAI compatible, text + multimodal)
  * [`list-models`](docs/voidnet_llm_list-models.md) - List publisher models (LLM, non-billable)

</details>
<!-- End Commands [operations] -->

<!-- Start Request Body Input [stdinpiping] -->
## Request Body Input

Commands that accept a request body take it three ways, with a clear priority chain: individual field flags (highest priority), the whole body as JSON via `--body`, and JSON piped on stdin (lowest priority). Later sources never override earlier ones; a body-bearing command prints its exact request schema with `--schema`.
<!-- End Request Body Input [stdinpiping] -->

<!-- Start Output Formats [output-formats] -->
## Output Formats

Every command supports a `--output-format` flag that controls how the response is rendered to stdout.

### Available formats

| Format | Flag | Description |
|--------|------|-------------|
| Pretty | `--output-format pretty` (default) | Aligned key-value pairs with color, nested indentation. Human-readable at a glance. |
| JSON | `--output-format json` | JSON output. Passthrough when the response is already JSON (preserves original field order and numeric precision). Falls back to typed marshaling otherwise. |
| YAML | `--output-format yaml` | YAML output via standard marshaling. |
| Table | `--output-format table` | Tabular output for array responses. |
| TOON | `--output-format toon` | [Token-Oriented Object Notation](https://github.com/toon-format/spec) — a compact, line-oriented format that typically uses 30–60% fewer tokens than JSON. Well-suited for piping responses into LLM prompts. |

```bash
# Default pretty output
voidnet publisher list-apps

# Machine-readable JSON
voidnet publisher list-apps --output-format json

# TOON for LLM-friendly compact output
voidnet publisher list-apps --output-format toon

# Pipe JSON to jq without using --output-format
voidnet publisher list-apps --output-format json | jq '.'
```

### jq filtering

Use `--jq` to filter or transform the response inline using a [jq](https://jqlang.org) expression. This always outputs JSON and overrides `--output-format`:

```bash
# Extract a single field
voidnet publisher list-apps --jq '.'

# Reshape with any jq program; --raw-output prints string results as plain text (like jq -r)
voidnet publisher list-apps --jq '.' --raw-output
```

### Color control

Use `--color` to control terminal colors:

| Value | Behavior |
|-------|----------|
| `auto` (default) | Color when stdout is a TTY, plain text otherwise |
| `always` | Always colorize |
| `never` | Never colorize |

The `NO_COLOR` and `FORCE_COLOR` environment variables are also respected.

### Streaming and pagination

When using `--all` (pagination) or streaming operations, output is written incrementally as items arrive:

| Format | Streaming behavior |
|--------|-------------------|
| `json` | One compact JSON object per line ([NDJSON](https://github.com/ndjson/ndjson-spec)) |
| `yaml` | YAML documents separated by `---` |
| `toon` | One TOON-encoded object per block, separated by blank lines |
| `pretty` (default) | Pretty-printed items separated by blank lines |
<!-- End Output Formats [output-formats] -->

<!-- Start Error Handling [errors] -->
## Error Handling

The CLI uses standard exit codes to indicate success or failure:

| Exit Code | Meaning |
|-----------|---------|
| `0` | Success |
| `1` | Runtime/API failure |
| `2` | Usage or input failure |
| `3` | Authentication or authorization failure |

On success, the response data is printed to **stdout** as JSON. On failure, error details are printed to **stderr**.

```bash
# Capture output and handle errors
voidnet publisher list-apps --output-format json > output.json 2> error.log
if [ $? -ne 0 ]; then
  echo "Error occurred, see error.log"
fi
```
This CLI uses unclassified error rendering outside agent mode: pretty and TOON print the API error text as received, while `--output-format json` and `--jq` emit the unclassified envelope (including the configure `_hint` for HTTP 401/403) plus `exit_code`. Agent mode always emits the classified JSON envelope with `exit_code`, `error_type`, optional `error_reason`, `message`, `hints`, and optional `status_code` — see [For AI agents](#for-ai-agents).
<!-- End Error Handling [errors] -->

<!-- Start Diagnostics [diagnostics] -->
## Diagnostics

The CLI includes two diagnostic flags available on all commands:

### Dry Run

Preview what would be sent without making any network calls:

```bash
voidnet publisher list-apps --dry-run
```

In human output modes, stdout is empty and the `[DRY-RUN]` block goes to stderr. It includes:
- HTTP method and URL
- Request headers (sensitive values redacted)
- Request body preview (sensitive fields redacted)

With `--output-format json`, or with a caller-explicit `--jq`, stderr is silent and stdout is NDJSON: one compact preview object per would-be request. The jq filter is not applied, and command-declared jq presets do not select the JSON protocol.

```json
{"dry_run":true,"request":{"method":"POST","url":"https://…","headers":{"Accept":["application/json"],…},"body":<JSON value | string | null>}}
```

JSON bodies remain structured; text bodies are strings; binary bodies are `"<bytes:N>"`; absent bodies are `null`. Headers retain all values as arrays, with credentials replaced by `[REDACTED]`. Dry-run never reads the OS keychain, but credentials supplied by flag, environment, or config file still appear redacted. The command exits successfully without contacting the API.

Local mutation commands emit one `{"dry_run":true,"local":true,"command":"…","message":"…"}` object in place of a preview; filter with `select(.request)` or `select(.local)`.

### Debug

Log request and response diagnostics while running normally:

```bash
voidnet publisher list-apps --debug
```

Debug output goes to stderr and includes:
- Request method, URL, headers, and body preview
- Response status, headers, and body preview
- Transport errors (if any)

The command still executes normally and produces its regular output on stdout.

### Flag Precedence

If both `--dry-run` and `--debug` are set, `--dry-run` takes precedence and no network calls are made.

### Security

Sensitive information is automatically redacted in diagnostic output:
- **Headers**: `Authorization`, `Cookie`, `Set-Cookie`, `X-API-Key`, and other security headers show `[REDACTED]`
- **Body**: JSON fields named `password`, `secret`, `token`, `api_key`, `client_secret`, etc. show `[REDACTED]`
- **Binary data**: binary media and canonical base64 strings are replaced with `<bytes:N>`
- **URL query**: credential-like query parameters are replaced with `[REDACTED]`

Diagnostic output should still be treated as potentially sensitive operational data.
<!-- End Diagnostics [diagnostics] -->

<!-- Placeholder for Future Speakeasy SDK Sections -->
