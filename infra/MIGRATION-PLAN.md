# Istio Test-Infra: GCP → AWS Migration Plan

## Problem & Approach

Istio's test infrastructure runs primarily on GCP (6 projects) plus Cloudflare.
Object storage has already moved to **Cloudflare R2**, and container registries are
moving to **GitHub Container Registry (GHCR)**. The remaining GCP resources (
Kubernetes (Prow CI/CD), identity, secrets, and signing, and bots ) must move to a **single AWS
account** (Equivalent to one GCP project).

---

## Target Platform Summary

| Layer | Today (GCP) | Target |
|---|---|---|
| Object storage | GCS buckets | **Cloudflare R2** (done) |
| Container registries | GCR / Artifact Registry | **GHCR** (in progress) |
| DNS | Cloud DNS + Cloudflare | **Cloudflare** (unchanged) |
| Kubernetes / CI | GKE (Prow) | **Amazon EKS** |
| VMs | Compute Engine | **EC2** |
| Secrets | Secret Manager | **AWS Secrets Manager** |
| Signing keys | Cloud KMS | **AWS KMS** (or keyless cosign) |
| Identity | Service Accounts + Workload Identity | **IAM Roles + IRSA / EKS Pod Identity** |
| Serverless | Cloud Run | **(decide: drop / Lambda / container)** |

---

## Resource Catalogue

A catalogue of our current GCP resources and what each is used for, grouped by project.

### `istio-testing` — Prow control plane + testing (largest project)

| Resource | Usage |
|---|---|
| GKE `prow` | Prow **control plane** (Deck/Hook/Tide/plank); schedules jobs onto build clusters |
| GKE `policy-bot` (+ node pool) | Hosts the Policy Bot (see below) |
| Spanner `istio-policy-bot` | Policy Bot primary datastore |
| Compute instance `elekto-web` (+ 2 SSH firewalls) | `elections.istio.io` VM (Elekto) |
| KMS `istio-cosign-keyring` | Cosign keyring — **likely an unused duplicate** of the prow-build one |
| Secret Manager IAM (x4) | Grants External Secrets access to secrets |
| Service Accounts (x7): `prow-control-plane`, `prow-deployer`, `testgrid-updater`, `kubernetes-external-secrets-sa`, `istio-policy-bot`, `istio-prow-test-job-default`, `istio-prow-test-job` (obsolete) | Workload identities for control-plane components and jobs |
| Registry `gcr.io/istio-testing` (backed by `artifacts.istio-testing.appspot.com`) | Development build / tooling images |
| GCS buckets (x18) | Test artifacts/results (`istio-prow`), build artifacts (`istio-build`), testgrid; rest are legacy (see `buckets.md`) |

### `istio-prow-build` — public build

| Resource | Usage |
|---|---|
| GKE `prow` (+ build/test node pools) | **Core build cluster** — most jobs run here |
| GKE `prow-arm` (+ ARM spot pool) | ARM jobs; separate only because GCP ARM is single-zone |
| KMS `istio-cosign-keyring` (+ key policy) | Cosign **release-artifact signing**; usable only by the release job |
| Secret Manager `release_secrets` | DockerHub token, GitHub PATs, Grafana token |
| Service Account `istio-prow-jobs` | Workload identity for build jobs |
| GCS buckets (x4) | Build artifacts |

### `istio-prow-private` — private build (sensitive)

| Resource | Usage |
|---|---|
| GKE `prow` (private) | **Private build cluster** (PSWG-restricted) |
| Artifact Registry `istio-prow-private` (`us-docker.pkg.dev`, + IAM policy) | Private container images |
| GCS `istio-build-private`, `istio-prerelease-private`, `istio-prow-private` | Private mirrors of the public build/prerelease/prow buckets |
| Project IAM members (x6) | PSWG-only access bindings |

### `istio-io` — releases + web

| Resource | Usage |
|---|---|
| GCS `istio-release` (prod), `istio-prerelease`, `fortio-data` | Official + prerelease release artifacts; legacy fortio data |
| GCS `istio-terraform` | Terraform state backend |
| Cloud DNS zone `istio.io` | Production DNS for istio.io |
| Cloud Run `redirector` | Slack invite link redirector |

### `istio-release` & `istio-prerelease-testing`

| Resource | Usage |
|---|---|
| Registry `gcr.io/istio-release` (backed by `artifacts.istio-release.appspot.com`) | **Production** release container images |
| Registry `gcr.io/istio-prerelease-testing` (backed by `artifacts.istio-prerelease-testing.appspot.com`) | Prerelease container images |
| GCS release buckets | Backing storage for the registries above |
| Project IAM members | Admin/owner bindings |

### Deprecated / legacy registries

| Resource | Usage |
|---|---|
| `gcr.io/istio-prow-build` (`artifacts.istio-prow-build.appspot.com`) | Old private proxy builds — deprecated |
| `gcr.io/istio-io` (`artifacts.istio-io.appspot.com`) | Old artifacts only — legacy |

### Cloudflare (`cf/cncf-istio`)

| Resource | Usage |
|---|---|
| DNS zone + records | Istio DNS |
| Registry-redirector Worker + ruleset | Redirects registry URLs so consumers are insulated from backend moves |

---

## Policy Bot (high level)

The Policy Bot enforces GitHub hygiene (labels, nags, stale-close, flake tracking) and
serves the eng.istio.io dashboard. Its data is **largely a rebuildable mirror of GitHub**.

| Dependency | Usage |
|---|---|
| GKE cluster | Runs the bot's server + cron-manager pods |
| Spanner | Primary state store (GitHub mirror) — **7 nodes, heavily oversized** |
| GCS blobstore | Reads Prow test artifacts (`istio-prow`) for flake/coverage analysis |
| BigQuery (GH Archive) | Historical activity backfill only; live webhooks already cover ongoing data |


---

## Security Boundaries (single-account model) — **emphasis**

GCP projects gave hard isolation "for free." In one AWS account this must be engineered:

- **Separate Networks** per cluster
- **EKS access entries**: the control-plane role is granted Kubernetes RBAC **only** in the
  clusters it is allowed to schedule into; the private cluster trusts only its dedicated
  scheduling path.
- **Resource policies + KMS key policies** isolate private artifacts, release secrets, and
  the cosign signing key so public-zone roles cannot reach them.

---

## Federated / Workload Identity for K8s Service Accounts — **emphasis**

Today the bridge between Kubernetes ServiceAccounts and cloud IAM is **GKE Workload
Identity**. We can do the same on AWS with EKS Pod Identity.

---
