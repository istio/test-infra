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
  ])

}
