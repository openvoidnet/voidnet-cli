#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)"
ACC="${VOIDNET_AUTH_BASE:-http://localhost:3020}"
NET="${VOIDNET_MGMT_BASE:-http://localhost:3000}"
PASS=0
FAIL=0
check() {
  local name="$1" want_code="$2" want_frag="$3" got="$4"
  local code frag
  code="$(echo "$got" | head -n 1)"
  frag="$(echo "$got" | tail -n +2)"
  if [ "$code" = "$want_code" ] && echo "$frag" | grep -q "$want_frag"; then
    PASS=$((PASS + 1))
    echo "[contract] PASS: $name (http=$code contains '$want_frag')"
  else
    FAIL=$((FAIL + 1))
    echo "[contract] FAIL: $name (want http=$want_code + '$want_frag', got http=$code body=$(echo "$frag" | head -c 200))" >&2
  fi
}
get() {
  local url="$1"
  local body code
  body="$(timeout 20 curl -s -m 15 -o /tmp/contract-body.txt -w "%{http_code}" "$url")"
  code="$body"
  echo "$code"
  cat /tmp/contract-body.txt
}
post() {
  local url="$1" data="$2"
  local code
  code="$(timeout 20 curl -s -m 15 -X POST "$url" -H 'Content-Type: application/json' -d "$data" -o /tmp/contract-body.txt -w "%{http_code}")"
  echo "$code"
  cat /tmp/contract-body.txt
}
echo "[contract] accounts=$ACC net=$NET"
echo "[contract] 1: authorize stranger bounces to login"
timeout 20 curl -s -m 15 --max-redirs 0 -D /tmp/contract-hdrs.txt -o /dev/null "$ACC/api/auth/oauth/authorize?response_type=code&client_id=voidnet-cli&redirect_uri=http://127.0.0.1:9/callback&scope=publisher_login&state=c1&code_challenge=E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM&code_challenge_method=S256" || true
HD="$(grep -i "^HTTP" /tmp/contract-hdrs.txt | tail -n 1 | awk '{print $2}')"
LOC="$(grep -i "^location:" /tmp/contract-hdrs.txt | tail -n 1)"
if [ "$HD" = "307" ] && echo "$LOC" | grep -q "/login?oauth=true"; then PASS=$((PASS+1)); echo "[contract] PASS: authorize stranger 307 to login"; else FAIL=$((FAIL+1)); echo "[contract] FAIL: authorize stranger bounce (http=$HD loc=$LOC)" >&2; fi
echo "[contract] 2: authorize unknown client"
R="$(post "$ACC/api/auth/oauth/authorize" '{}' || true)"
echo "[contract] 2: authorize is GET-only path (skip, covered by 3)"
echo "[contract] 3: authorize missing PKCE for public client"
B="$(timeout 20 curl -s -m 15 "$ACC/api/auth/oauth/authorize?response_type=code&client_id=voidnet-cli&redirect_uri=http://127.0.0.1:9/callback&scope=publisher_login&state=c1" -o /tmp/contract-body.txt -w "%{http_code}")"
check "authorize public without PKCE" "400" "code_challenge" "$B
$(cat /tmp/contract-body.txt)"
echo "[contract] 4: authorize unknown client"
B="$(timeout 20 curl -s -m 15 "$ACC/api/auth/oauth/authorize?response_type=code&client_id=nope&redirect_uri=http://127.0.0.1:9/callback" -o /tmp/contract-body.txt -w "%{http_code}")"
check "authorize unknown client" "401" "Invalid client" "$B
$(cat /tmp/contract-body.txt)"
echo "[contract] 5: token unknown code"
R="$(post "$ACC/api/auth/oauth/token" '{"grant_type":"authorization_code","client_id":"voidnet-cli","code":"deadbeef","redirect_uri":"http://127.0.0.1:9/callback","code_verifier":"x"}')"
check "token unknown code" "400" "invalid_grant" "$R"
echo "[contract] 6: token bad grant type"
R="$(post "$ACC/api/auth/oauth/token" '{"grant_type":"password","client_id":"voidnet-cli"}')"
check "token bad grant" "400" "Unsupported grant type" "$R"
echo "[contract] 7: consent-data without session"
R="$(get "$ACC/api/auth/oauth/consent-data?client_id=voidnet-cli")"
check "consent-data no session" "401" "Sign-in required" "$R"
echo "[contract] 8: consent approve without session"
R="$(post "$ACC/api/auth/oauth/consent" '{"client_id":"voidnet-cli","redirect_uri":"http://127.0.0.1:9/callback"}')"
check "consent no session" "401" "Sign-in required" "$R"
echo "[contract] 9: exchange missing token"
R="$(post "$NET/api/publisher/v1/auth/exchange" '{}')"
check "exchange missing token" "400" "missing_token" "$R"
echo "[contract] 10: exchange garbage token"
R="$(post "$NET/api/publisher/v1/auth/exchange" '{"accounts_token":"garbage"}')"
check "exchange invalid token" "401" "invalid_token" "$R"
echo "[contract] 11: revoke-self garbage token"
C="$(timeout 20 curl -s -m 15 -X POST "$NET/api/publisher/v1/auth/revoke-self" -H "Authorization: Bearer garbage" -o /tmp/contract-body.txt -w "%{http_code}")"
check "revoke-self invalid" "401" "publisher token" "$C
$(cat /tmp/contract-body.txt)"
echo "[contract] 12: caller profile without bearer"
C="$(timeout 20 curl -s -m 15 "$NET/api/publisher/v1/auth/me" -o /tmp/contract-body.txt -w "%{http_code}")"
check "profile missing bearer" "400" "missing_token" "$C
$(cat /tmp/contract-body.txt)"
echo "[contract] 13: caller profile garbage token"
C="$(timeout 20 curl -s -m 15 "$NET/api/publisher/v1/auth/me" -H "Authorization: Bearer garbage" -o /tmp/contract-body.txt -w "%{http_code}")"
check "profile invalid token" "401" "invalid_token" "$C
$(cat /tmp/contract-body.txt)"
echo "[contract] result: PASS=$PASS FAIL=$FAIL"
if [ "$FAIL" -ne 0 ]; then echo "[contract] CONTRACT BROKEN" >&2; exit 1; fi
echo "[contract] CONTRACT OK"
