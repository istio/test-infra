#!/usr/bin/env bash
#
# sync-secrets-from-gcp.sh
#
# Copies the latest enabled version of each secret from Google Secret Manager
# into the matching AWS Secrets Manager secret. Secret names are identical on
# both sides; only the source GCP project differs.
#
# This is a manual, one-shot convenience for keeping AWS secret values current
# until the in-cluster rotator owns them. It only writes new secret *values* to
# secrets that already exist in AWS (created by Terraform); it never creates,
# renames, or deletes secrets.
#
# Usage:
#   ./sync-secrets-from-gcp.sh [--dry-run] [--region REGION] [--only NAME]...
#
# Requirements: gcloud (authenticated), aws (authenticated), jq not required.

set -euo pipefail

AWS_REGION="${AWS_REGION:-us-west-2}"
DRY_RUN=false
ONLY=()

usage() {
  sed -n '2,20p' "$0" | sed 's/^# \{0,1\}//'
  exit "${1:-0}"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY_RUN=true; shift ;;
    --region)  AWS_REGION="$2"; shift 2 ;;
    --only)    ONLY+=("$2"); shift 2 ;;
    -h|--help) usage 0 ;;
    *) echo "unknown argument: $1" >&2; usage 1 ;;
  esac
done

# secret name -> source GCP project. Names are reused verbatim as the AWS
# secret id and the GCP secret id.
read -r -d '' SECRET_MAP <<'EOF' || true
istio-prow-build  release_docker_istio
istio-prow-build  release_github_istio-release
istio-prow-build  release_grafana_istio
istio-prow-build  github-read_github_read
istio-prow-build  github_istio-testing_pusher
istio-testing     cf_r2_admin_token
istio-testing     cf_r2_public_buckets_ro_credentials
istio-testing     cf_r2_istio-build_credentials
istio-testing     cf_r2_istio-build-private_credentials
istio-testing     cf_r2_istio-prerelease_credentials
istio-testing     cf_r2_istio-prerelease-private_credentials
istio-testing     cf_r2_istio-prow_credentials
istio-testing     cf_r2_istio-prow-private_credentials
istio-testing     cf_r2_istio-testgrid_credentials
istio-testing     cf_r2_istio-release_credentials
EOF

command -v gcloud >/dev/null 2>&1 || { echo "error: gcloud not found on PATH" >&2; exit 1; }
command -v aws    >/dev/null 2>&1 || { echo "error: aws not found on PATH" >&2; exit 1; }

want() {
  [[ ${#ONLY[@]} -eq 0 ]] && return 0
  local n="$1"
  for o in "${ONLY[@]}"; do [[ "$o" == "$n" ]] && return 0; done
  return 1
}

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

synced=0
skipped=0
failed=0

while read -r project name; do
  [[ -z "${project:-}" || -z "${name:-}" ]] && continue
  want "$name" || continue

  payload="$TMP_DIR/$name"

  if ! gcloud secrets versions access latest \
        --project "$project" --secret "$name" \
        --out-file "$payload" >/dev/null 2>&1; then
    echo "SKIP  $name (no readable version in gcp project $project)" >&2
    skipped=$((skipped + 1))
    continue
  fi

  if [[ ! -s "$payload" ]]; then
    echo "SKIP  $name (empty value in gcp)" >&2
    skipped=$((skipped + 1))
    continue
  fi

  if ! aws secretsmanager describe-secret \
        --region "$AWS_REGION" --secret-id "$name" >/dev/null 2>&1; then
    echo "FAIL  $name (no matching secret in aws region $AWS_REGION)" >&2
    failed=$((failed + 1))
    continue
  fi

  if $DRY_RUN; then
    echo "DRYRUN $name ($(wc -c <"$payload" | tr -d ' ') bytes from $project)"
    synced=$((synced + 1))
    continue
  fi

  if aws secretsmanager put-secret-value \
        --region "$AWS_REGION" --secret-id "$name" \
        --secret-string "file://$payload" >/dev/null; then
    echo "SYNC  $name (from $project)"
    synced=$((synced + 1))
  else
    echo "FAIL  $name (aws put-secret-value failed)" >&2
    failed=$((failed + 1))
  fi
done <<<"$SECRET_MAP"

echo "----"
echo "synced=$synced skipped=$skipped failed=$failed"
[[ "$failed" -eq 0 ]]
