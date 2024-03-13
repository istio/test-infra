locals {
  # Members with view access to private infrastructure.
  # Anyone here MUST be in the PSWG
  # Comment indicates github username
  private_infra_viewers = toset([
    "howardjohn@google.com",     # howardjohn
    "johnbhoward96@gmail.com",   # howardjohn
    "ketihmattix@microsoft.com", # keithmattix
    "daniel.hawton@solo.io",     #dhawton
    "paul@tetrate.io",           # pmerrison 
  ])

  terraform_infra_admins = toset([
    "howardjohn@google.com",   # howardjohn
    "johnbhoward96@gmail.com", # howardjohn
  ])

}
