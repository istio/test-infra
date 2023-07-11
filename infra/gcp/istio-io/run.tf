# This hosts the Istio Slack invitation service.
# See https://docs.google.com/document/d/1Pf6vS5SiYuSD55atbosoSHktIi1N08p-P480buzCrDk/ for more info
# The invite link needs to be updated every 400 invites. TODO: find a way to automate this?
resource "google_cloud_run_v2_service" "redirector" {
  location = "us-central1"
  name     = "redirector"
  project  = local.project_id
  template {
    max_instance_request_concurrency = 80
    revision                         = null
    service_account                  = "589046318233-compute@developer.gserviceaccount.com"
    session_affinity                 = false
    timeout                          = "900s"
    containers {
      # Built from https://github.com/ahmetb/serverless-url-redirect
      image = "gcr.io/istio-testing/redirector"
      env {
        name  = "REDIRECT_URL"
        value = "https://join.slack.com/t/istio/shared_invite/zt-1z9yjkym8-HTFsoVo4Zom7LOwBXxjc2Q"
      }
      resources {
        cpu_idle = true
        limits = {
          cpu    = "1000m"
          memory = "256Mi"
        }
      }
    }
  }
}
