#!/usr/bin/env bash
set -euo pipefail
EXP="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
echo "[exp-cli] lint: gateway spec"
speakeasy lint openapi -s "$EXP/specs/gateway.json"
echo "[exp-cli] lint: publisher spec"
speakeasy lint openapi -s "$EXP/specs/publisher.json"
echo "[exp-cli] lint: workflow + target config"
(cd "$EXP/.speakeasy" && speakeasy lint config)
echo "[exp-cli] lint OK"
