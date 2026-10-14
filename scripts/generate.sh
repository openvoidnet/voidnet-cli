#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)"
EXP="$ROOT/packages/exp/cli"
export GOTOOLCHAIN=auto
echo "[exp-cli] auth gate: checking speakeasy authentication"
if [ -z "${SPEAKEASY_API_KEY:-}" ] && ! timeout 30 speakeasy status --output console >/dev/null 2>&1; then
  echo "[exp-cli] ERROR: speakeasy is not authenticated." >&2
  echo "[exp-cli] Run 'speakeasy auth login' once, or export SPEAKEASY_API_KEY, then re-run $0." >&2
  exit 1
fi
echo "[exp-cli] auth OK"
echo "[exp-cli] run: merge + overlay + CLI generation (target voidnet-cli)"
(cd "$EXP" && speakeasy run -t voidnet-cli --skip-upload-spec --output console)
echo "[exp-cli] run OK"
echo "[exp-cli] wiring hand code (deterministic patches, fail loud on drift)"
for p in "$EXP"/scripts/wire-*.patch; do
  echo "[exp-cli] patch: $(basename "$p")"
  (cd "$EXP" && patch --dry-run -p1 < "$p") || { echo "[exp-cli] ERROR: $(basename "$p") no longer applies — template changed. Update the patch." >&2; exit 1; }
  (cd "$EXP" && patch -s -p1 < "$p" && gofmt -w internal/cli/root.go internal/output/output.go)
done
echo "[exp-cli] hand code wired"
echo "[exp-cli] discovering generated Cobra main (excluding gendocs)"
CLI_NAME="$(ls "$EXP/cmd" | grep -v '^gendocs$' | head -n 1)"
if [ -z "$CLI_NAME" ]; then
  echo "[exp-cli] ERROR: no generated command found under $EXP/cmd (expected <cliName> + gendocs)." >&2
  exit 1
fi
echo "[exp-cli] cliName=$CLI_NAME"
echo "[exp-cli] deps: go mod tidy"
(cd "$EXP" && go mod tidy)
echo "[exp-cli] build: go build -o voidnet ./cmd/$CLI_NAME"
(cd "$EXP" && go build -o voidnet "./cmd/$CLI_NAME")
echo "[exp-cli] smoke: ./voidnet-cli-exp --help"
"$EXP/voidnet" --help > /tmp/exp-cli-help.txt
for g in login logout publisher mcp llm health whoami; do
  grep -q "^  $g " /tmp/exp-cli-help.txt || { echo "[exp-cli] ERROR: main command '$g' missing from front page." >&2; exit 1; }
done
"$EXP/voidnet" help -a > /tmp/exp-cli-help-all.txt
for g in auth configure explore completion version; do
  grep -q "^  $g" /tmp/exp-cli-help-all.txt || { echo "[exp-cli] ERROR: utility '$g' missing from help -a." >&2; exit 1; }
done
echo "[exp-cli] surface gate OK: front page curated, utilities under help -a"
echo "[exp-cli] generate OK: binary at $EXP/voidnet"
