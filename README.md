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
