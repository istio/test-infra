locals {
  # Members with view access to private infrastructure.
  # Anyone here MUST be in the PSWG
  # Comment indicates github username
  private_infra_viewers = toset([
    "john.howard@solo.io",       # howardjohn
    "johnbhoward96@gmail.com",   # howardjohn
    "keithmattix@microsoft.com", # keithmattix
    "daniel.hawton@solo.io",     # dhawton
    "paul@tetrate.io",           # pmerrison
  ])

  terraform_infra_admins = toset([
    "john.howard@solo.io",     # howardjohn
    "johnbhoward96@gmail.com", # howardjohn
    "daniel.hawton@solo.io",   # dhawton
    "keithmattix2@gmail.com",  # keithmattix
  ])

  kind_node_images = {
    "kindest/node:v1.30.13" : "v1.30.13",
    "kindest/node:v1.31.9" : "v1.31.0",
    "kindest/node:v1.32.5" : "v1.32.5",
    "kindest/node:v1.33.1" : "v1.33.1",
  }
}
