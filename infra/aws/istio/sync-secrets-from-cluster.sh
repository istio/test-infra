#!/usr/bin/env bash
#
# sync-secrets-from-cluster.sh
#
# Copies the hand-created Prow control-plane secrets that live only as plain
# Kubernetes Secrets on the source GKE `prow` cluster into the matching AWS
# Secrets Manager secrets. These are NOT managed by any operator on GKE, so the
# live cluster is their source of truth.
#
# Single-key secrets are stored in AWS as the raw value of that key. The one
# multi-key secret (deck-oauth-proxy) is serialized to a JSON object so the AWS
# value carries all keys.
#
# It only writes values to AWS secrets that already exist (created by
# Terraform); it never creates, renames, or deletes secrets.
#
# Usage:
#   ./sync-secrets-from-cluster.sh [--dry-run] [--region REGION] [--context CTX]
#
# Requirements: kubectl (with access to the source cluster), aws (authenticated).

set -euo pipefail

AWS_REGION="${AWS_REGION:-us-west-2}"
KUBE_CONTEXT="${KUBE_CONTEXT:-gke_istio-testing_us-west1-a_prow}"
SRC_NS="default"
DRY_RUN=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY_RUN=true; shift ;;
    --region)  AWS_REGION="$2"; shift 2 ;;
    --context) KUBE_CONTEXT="$2"; shift 2 ;;
    -h|--help) sed -n '2,21p' "$0" | sed 's/^# \{0,1\}//'; exit 0 ;;
    *) echo "unknown argument: $1" >&2; exit 1 ;;
  esac
done

command -v kubectl >/dev/null 2>&1 || { echo "error: kubectl not found on PATH" >&2; exit 1; }
command -v aws     >/dev/null 2>&1 || { echo "error: aws not found on PATH" >&2; exit 1; }

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

# Pull one key out of a source Secret, base64-decoded, into a temp file.
k8s_key() {
  local secret="$1" key="$2" out="$3"
  kubectl --context "$KUBE_CONTEXT" -n "$SRC_NS" get secret "$secret" \
    -o "jsonpath={.data.$key}" | base64 -d >"$out"
}

# Write a temp file's contents into an existing AWS secret.
push() {
  local aws_id="$1" payload="$2"
  if ! aws secretsmanager describe-secret --region "$AWS_REGION" --secret-id "$aws_id" >/dev/null 2>&1; then
    echo "FAIL  $aws_id (no matching secret in aws region $AWS_REGION)" >&2
    return 1
  fi
  if $DRY_RUN; then
    echo "DRYRUN $aws_id ($(wc -c <"$payload" | tr -d ' ') bytes)"
    return 0
  fi
  aws secretsmanager put-secret-value --region "$AWS_REGION" \
    --secret-id "$aws_id" --secret-string "file://$payload" >/dev/null
  echo "SYNC  $aws_id"
}

# --- oauth-token (single key: oauth) -> raw string in AWS oauth_token ---
k8s_key oauth-token oauth "$TMP_DIR/oauth_token"
push oauth_token "$TMP_DIR/oauth_token"

# --- hmac-token (single key: hmac) -> raw string in AWS hmac_token ---
k8s_key hmac-token hmac "$TMP_DIR/hmac_token"
push hmac_token "$TMP_DIR/hmac_token"

# --- cookie (single key: secret) -> raw string in AWS cookie ---
k8s_key cookie secret "$TMP_DIR/cookie"
push cookie "$TMP_DIR/cookie"

# --- github-oauth-config (single key: secret) -> raw string ---
k8s_key github-oauth-config secret "$TMP_DIR/github-oauth-config"
push github-oauth-config "$TMP_DIR/github-oauth-config"

# --- github-oauth-config-private (single key: secret) -> raw string ---
k8s_key github-oauth-config-private secret "$TMP_DIR/github-oauth-config-private"
push github-oauth-config-private "$TMP_DIR/github-oauth-config-private"

# --- slack-token (single key: token) -> raw string in AWS slack_token ---
k8s_key slack-token token "$TMP_DIR/slack_token"
push slack_token "$TMP_DIR/slack_token"

# --- deck-oauth-proxy (multi key) -> JSON {clientID,clientSecret,cookieSecret} ---
kubectl --context "$KUBE_CONTEXT" -n "$SRC_NS" get secret deck-oauth-proxy -o json \
  | python3 -c '
import base64, json, sys
data = json.load(sys.stdin)["data"]
out = {k: base64.b64decode(v).decode("utf-8") for k, v in data.items()}
sys.stdout.write(json.dumps(out))
' >"$TMP_DIR/deck-oauth-proxy"
push deck-oauth-proxy "$TMP_DIR/deck-oauth-proxy"

echo "----"
echo "done (region=$AWS_REGION context=$KUBE_CONTEXT dry_run=$DRY_RUN)"
