# Auth domain configuration examples (`AuthtypesAuthDomainConfig` in SigNoz OpenAPI).
# Copy the resource block that matches your IdP. `ssoType`: `saml`, `oidc`, `google_auth`, or `email_password`.
# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.

resource "signoz_auth_domain" "saml" {
  name = "saml.example.com" # required (Terraform)

  config = jsonencode({                                                    # optional (Terraform)
    ssoEnabled = true                                                      # required (SigNoz API)
    ssoType    = "saml"                                                    # required (SigNoz API)
    samlConfig = {                                                         # required when ssoType is saml
      samlEntity                      = "https://idp.example.com/metadata" # required (SigNoz API)
      samlIdp                         = "https://idp.example.com/sso/saml" # required (SigNoz API)
      samlCert                        = "MIIC...base64-certificate..."     # required (SigNoz API)
      insecureSkipAuthNRequestsSigned = false                              # optional
      attributeMapping = {                                                 # optional
        email  = "email"                                                   # optional (per-key)
        name   = "name"                                                    # optional (per-key)
        groups = "groups"                                                  # optional (per-key)
        role   = "role"                                                    # optional (per-key)
      }
    }
    roleMapping = {          # optional
      defaultRole = "VIEWER" # optional
      groupMappings = {      # optional
        "platform-admins" = "ADMIN"
        "developers"      = "EDITOR"
      }
      useRoleAttribute = false # optional
    }
  })
}

resource "signoz_auth_domain" "oidc" {
  name = "oidc.example.com" # required (Terraform)

  config = jsonencode({                                                      # optional (Terraform)
    ssoEnabled = true                                                        # required (SigNoz API)
    ssoType    = "oidc"                                                      # required (SigNoz API)
    oidcConfig = {                                                           # required when ssoType is oidc
      issuer                    = "https://login.example.com/realms/myrealm" # required (SigNoz API)
      issuerAlias               = ""                                         # optional
      clientId                  = "my-oidc-client-id"                        # required (SigNoz API)
      clientSecret              = "my-oidc-client-secret"                    # required (SigNoz API)
      getUserInfo               = true                                       # optional
      insecureSkipEmailVerified = false                                      # optional
      claimMapping = {                                                       # optional
        email  = "email"                                                     # optional (per-key)
        name   = "name"                                                      # optional (per-key)
        groups = "groups"                                                    # optional (per-key)
        role   = "roles"                                                     # optional (per-key)
      }
    }
    roleMapping = {          # optional
      defaultRole = "VIEWER" # optional
      groupMappings = {      # optional
        "platform-admins" = "ADMIN"
        "developers"      = "EDITOR"
      }
      useRoleAttribute = false # optional
    }
  })
}

resource "signoz_auth_domain" "google" {
  name = "google.example.com" # required (Terraform)

  # optional (Terraform)
  config = jsonencode({
    ssoEnabled = true                                                                      # required (SigNoz API)
    ssoType    = "google_auth"                                                             # required (SigNoz API)
    googleAuthConfig = {                                                                   # required when ssoType is google_auth
      clientId                       = "123456789.apps.googleusercontent.com"              # required (SigNoz API)
      clientSecret                   = "my-google-client-secret"                           # required (SigNoz API)
      redirectURI                    = "https://signoz.example.com/api/v1/callback/google" # required (SigNoz API)
      fetchGroups                    = true                                                # optional
      fetchTransitiveGroupMembership = false                                               # optional
      insecureSkipEmailVerified      = false                                               # optional
      # optional (required when fetchGroups is true)
      serviceAccountJson = "{\"type\":\"service_account\",\"project_id\":\"my-project\",\"private_key_id\":\"key-id\",\"private_key\":\"-----BEGIN PRIVATE KEY-----\\n...\\n-----END PRIVATE KEY-----\\n\",\"client_email\":\"signoz@my-project.iam.gserviceaccount.com\",\"client_id\":\"123456789\",\"auth_uri\":\"https://accounts.google.com/o/oauth2/auth\",\"token_uri\":\"https://oauth2.googleapis.com/token\"}"
      domainToAdminEmail = {                # optional (required when fetchGroups is true)
        "example.com" = "admin@example.com" # required (per-key when domainToAdminEmail is set)
      }
    }
    roleMapping = {          # optional
      defaultRole = "VIEWER" # optional
      groupMappings = {      # optional
        "platform-admins" = "ADMIN"
        "developers"      = "EDITOR"
      }
      useRoleAttribute = false # optional
    }
  })
}

resource "signoz_auth_domain" "password_only" {
  name = "password.example.com" # required (Terraform)

  config = jsonencode({           # optional (Terraform)
    ssoEnabled = false            # required (SigNoz API)
    ssoType    = "email_password" # required (SigNoz API)
  })
}

resource "signoz_auth_domain" "oidc_role_claim" {
  name = "oidc-role-claim.example.com" # required (Terraform)

  config = jsonencode({                                                      # optional (Terraform)
    ssoEnabled = true                                                        # required (SigNoz API)
    ssoType    = "oidc"                                                      # required (SigNoz API)
    oidcConfig = {                                                           # required when ssoType is oidc
      issuer                    = "https://login.example.com/realms/myrealm" # required (SigNoz API)
      clientId                  = "my-oidc-client-id"                        # required (SigNoz API)
      clientSecret              = "my-oidc-client-secret"                    # required (SigNoz API)
      getUserInfo               = true                                       # optional
      insecureSkipEmailVerified = false                                      # optional
      claimMapping = {                                                       # optional
        email = "email"                                                      # optional (per-key)
        name  = "name"                                                       # optional (per-key)
        role  = "roles"                                                      # optional (per-key)
      }
    }
    roleMapping = {               # optional
      defaultRole      = "VIEWER" # optional
      useRoleAttribute = true     # optional
      groupMappings    = {}       # optional
    }
  })
}
