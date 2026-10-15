#!/usr/bin/env bash
# CLI behavior matrix: asserts the eight structural defects stay dead.
# Needs local accounts (:3020) + net (:3000).
# Positive authenticated cases need a live read-scoped token:
#   VOIDNET_MATRIX_TOKEN=vnp-pub-... ./scripts/check-cli-matrix.sh
# Without it, positive cases report SKIP (loud, not silent).
set -euo pipefail
EXP="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="$EXP/voidnet"
SRV=(--server 1)
PASS=0
FAIL=0
SKIP=0
pass() { PASS=$((PASS + 1)); echo "[matrix] PASS: $1"; }
fail() { FAIL=$((FAIL + 1)); echo "[matrix] FAIL: $1" >&2; }
skip() { SKIP=$((SKIP + 1)); echo "[matrix] SKIP: $1 (no VOIDNET_MATRIX_TOKEN)"; }
expect_contains() {
  local name="$1" want="$2"; shift 2
  local out
  out="$("$@" 2>&1)" || true
  if echo "$out" | grep -q "$want"; then pass "$name"; else fail "$name (missing '$want' in: $(echo "$out" | head -c 160))"; fi
}
expect_ok() {
  local name="$1"; shift
  if "$@" >/dev/null 2>&1; then pass "$name"; else fail "$name (nonzero exit)"; fi
}
[ -x "$BIN" ] || { echo "[matrix] ERROR: binary missing at $BIN — run generate.sh first." >&2; exit 1; }
echo "[matrix] E3: unknown flag rejected without execution"
expect_contains "unknown flag rejected" "unknown flag: --nope" "$BIN" publisher list-apps --nope "${SRV[@]}"
echo "[matrix] E3: bad command rejected"
expect_contains "bad command rejected" 'unknown command "frobnicate"' "$BIN" frobnicate
echo "[matrix] E3: bare group shows help, not unknown-command"
expect_contains "bare group help" "voidnet publisher" "$BIN" publisher
echo "[matrix] E3: global -o works before and after command"
A="$("$BIN" -o json publisher list-apps "${SRV[@]}" 2>&1 | head -c 60)" || true
B="$("$BIN" publisher list-apps "${SRV[@]}" -o json 2>&1 | head -c 60)" || true
if [ "$A" = "$B" ]; then pass "global flag position-independent"; else fail "global flag position differs"; fi
echo "[matrix] E2: dead token names login remedy (pretty)"
expect_contains "401 remedy pretty" "voidnet login" env VOIDNET_PUBLISHER_TOKEN="vnp-pub-deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef" "$BIN" publisher list-apps "${SRV[@]}"
echo "[matrix] E2: dead token names login remedy (json)"
expect_contains "401 remedy json" "voidnet login" env VOIDNET_PUBLISHER_TOKEN="vnp-pub-deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef" "$BIN" publisher list-apps "${SRV[@]}" -o json
echo "[matrix] E1: garbage --token rejected without storing"
expect_contains "garbage token rejected" "must start with vnp-pub-" "$BIN" login --token garbage --mgmt-base http://localhost:3000
echo "[matrix] E1: dead --token rejected after API verification"
expect_contains "dead token rejected" "rejected by publisher API, nothing stored" "$BIN" login --token vnp-pub-deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef --mgmt-base http://localhost:3000
echo "[matrix] E6: success stdout is pure JSON"
if "$BIN" health check --server 1 -o json 2>/dev/null | python3 -c "import json,sys; json.load(sys.stdin)" 2>/dev/null; then pass "stdout pure JSON"; else fail "stdout not pure JSON"; fi
if [ -n "${VOIDNET_MATRIX_TOKEN:-}" ]; then
  echo "[matrix] E2: 403 names scope remedy (live read-only token)"
  expect_contains "403 scope remedy" "keys:rotate" env VOIDNET_PUBLISHER_TOKEN="$VOIDNET_MATRIX_TOKEN" "$BIN" publisher rotate-key --id 00000000-0000-0000-0000-000000000000 "${SRV[@]}"
  echo "[matrix] note: logout revocation proven manually (Phase 4); excluded here because it would consume the supplied matrix token"
else
  skip "403 scope remedy (live)"
  echo "[matrix] note: logout revocation proven manually (Phase 4); excluded here because it would consume the supplied matrix token"
fi
echo "[matrix] result: PASS=$PASS FAIL=$FAIL SKIP=$SKIP"
[ "$FAIL" -eq 0 ] || { echo "[matrix] MATRIX BROKEN" >&2; exit 1; }
echo "[matrix] MATRIX OK"
