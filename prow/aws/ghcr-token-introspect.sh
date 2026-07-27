#!/usr/bin/env bash
#
# Introspect the ghcr.io push GitHub App installation to disambiguate a
# "config error" (App/installation lacks a permission) from a "setup error"
# (permission is granted, but org policy / package settings block the push).
#
# It reads the App private key (PEM) straight from the k8s Secret that ESO syncs
# on the prow control-plane cluster, signs a short-lived App JWT locally, and
# asks GitHub for:
#   1. the installation's declared permissions (GET /app/installations/{id})
#   2. a freshly minted installation token's permissions + expires_at
#      (POST /app/installations/{id}/access_tokens)
#
# The `permissions.packages` value is the authoritative answer:
#   packages: write   -> config is fine; a push denial is an org/package-policy
#                        (setup) problem, not a token problem.
#   packages: read /
#   packages: absent  -> config problem: grant the App organization "Packages:
#                        write" and re-approve the installation.
#
# Nothing is written anywhere; all GitHub calls are read-only (the POST only
# mints a throwaway token). The token value itself is never printed.
#
# Usage:
#   ./ghcr-token-introspect.sh
#
# Env overrides (sensible defaults for the istio setup):
#   KUBE_CONTEXT   kube context for the prow cluster
#   NAMESPACE      namespace holding the PEM + generator (default: ghcr-push)
#   PEM_SECRET     Secret name holding the PEM     (default: ghcr-push-app-private-key)
#   PEM_KEY        data key within that Secret     (default: privateKey)
#   GENERATOR      GithubAccessToken name to read appID/installID from
#                                                  (default: ghcr-pusher)
#   APP_ID         override App ID    (else read from the generator)
#   INSTALL_ID     override Install ID (else read from the generator)
set -euo pipefail

KUBE_CONTEXT="${KUBE_CONTEXT:-arn:aws:eks:us-west-2:678412441677:cluster/prow}"
NAMESPACE="${NAMESPACE:-ghcr-push}"
PEM_SECRET="${PEM_SECRET:-ghcr-push-app-private-key}"
PEM_KEY="${PEM_KEY:-privateKey}"
GENERATOR="${GENERATOR:-ghcr-pusher}"

for bin in kubectl openssl curl jq; do
  command -v "$bin" >/dev/null 2>&1 || { echo "error: '$bin' not found on PATH" >&2; exit 1; }
done

kctl() { kubectl --context "$KUBE_CONTEXT" -n "$NAMESPACE" "$@"; }

# --- resolve App ID / Install ID -------------------------------------------
APP_ID="${APP_ID:-$(kctl get githubaccesstoken "$GENERATOR" -o jsonpath='{.spec.appID}' 2>/dev/null || true)}"
INSTALL_ID="${INSTALL_ID:-$(kctl get githubaccesstoken "$GENERATOR" -o jsonpath='{.spec.installID}' 2>/dev/null || true)}"
if [[ -z "${APP_ID}" || -z "${INSTALL_ID}" ]]; then
  echo "error: could not determine APP_ID/INSTALL_ID (set them via env or check generator '$GENERATOR' in $NAMESPACE)" >&2
  exit 1
fi

# --- pull the PEM from the k8s Secret --------------------------------------
PEM="$(kctl get secret "$PEM_SECRET" -o jsonpath="{.data.${PEM_KEY}}" 2>/dev/null | base64 -d || true)"
if [[ -z "${PEM}" ]]; then
  echo "error: could not read PEM from secret '$PEM_SECRET' key '$PEM_KEY' in $NAMESPACE" >&2
  exit 1
fi
if ! printf '%s' "$PEM" | openssl rsa -noout -check >/dev/null 2>&1; then
  echo "error: value in $PEM_SECRET/$PEM_KEY is not a valid RSA private key" >&2
  exit 1
fi

# --- build a short-lived (9 min) App JWT -----------------------------------
b64url() { openssl base64 -A | tr '+/' '-_' | tr -d '='; }
now=$(date +%s); iat=$((now - 60)); exp=$((now + 540))
header="$(printf '{"alg":"RS256","typ":"JWT"}' | b64url)"
payload="$(printf '{"iat":%d,"exp":%d,"iss":"%s"}' "$iat" "$exp" "$APP_ID" | b64url)"
signature="$(printf '%s' "${header}.${payload}" | openssl dgst -sha256 -sign <(printf '%s' "$PEM") -binary | b64url)"
JWT="${header}.${payload}.${signature}"
echo $JWT

api() {
  # $1 method, $2 path
  curl -sS -X "$1" \
    -H "Authorization: Bearer ${JWT}" \
    -H "Accept: application/vnd.github+json" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "https://api.github.com${2}"
}

echo "== App ${APP_ID}, installation ${INSTALL_ID} =="
echo
echo "== Installation (declared) permissions =="
api GET "/app/installations/${INSTALL_ID}" \
  | jq '{account: .account.login, target_type, repository_selection, permissions}'

echo
echo "== Freshly minted installation token =="
api POST "/app/installations/${INSTALL_ID}/access_tokens" \
  | jq '{expires_at, repository_selection, permissions}'

echo
pkgs="$(api POST "/app/installations/${INSTALL_ID}/access_tokens" | jq -r '.permissions.packages // "none"')"
case "$pkgs" in
  write)
    echo "VERDICT: packages=write is GRANTED -> config OK. A push denial is a"
    echo "         setup/org-policy problem (org 'Package creation' restriction or"
    echo "         the package needs to be pre-created/linked), not a token issue." ;;
  read)
    echo "VERDICT: packages=read only -> CONFIG problem. Grant the App organization"
    echo "         'Packages: write' and re-approve the installation." ;;
  *)
    echo "VERDICT: packages permission is ABSENT -> CONFIG problem. The App has no"
    echo "         packages permission; add organization 'Packages: write' and"
    echo "         re-approve the installation." ;;
esac
