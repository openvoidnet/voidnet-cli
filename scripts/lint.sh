#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)"
EXP="$ROOT/packages/exp/cli"
echo "[exp-cli] lint: gateway spec"
speakeasy lint openapi -s "$ROOT/apps/gateway/openapi/openapi.json"
echo "[exp-cli] lint: publisher spec"
speakeasy lint openapi -s "$ROOT/apps/net/openapi/openapi.json"
echo "[exp-cli] lint: workflow + target config"
(cd "$EXP/.speakeasy" && speakeasy lint config)
echo "[exp-cli] lint OK"
