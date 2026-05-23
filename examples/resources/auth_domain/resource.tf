# Auth domain configuration examples (`AuthtypesAuthDomainConfig` in SigNoz OpenAPI).
# Copy the resource block that matches your IdP. `ssoType`: `saml`, `oidc`, `google_auth`, or `email_password`.

resource "signoz_auth_domain" "saml" {
  name = "saml.example.com"

  config = jsonencode({
    ssoEnabled = true
    ssoType    = "saml"
    samlConfig = {
      samlEntity                      = "https://idp.example.com/metadata"
      samlIdp                         = "https://idp.example.com/sso/saml"
      samlCert                        = "MIIC...base64-certificate..."
      insecureSkipAuthNRequestsSigned = false
      attributeMapping = {
        email  = "email"
        name   = "name"
        groups = "groups"
        role   = "role"
      }
    }
    roleMapping = {
      defaultRole = "VIEWER"
      groupMappings = {
        "platform-admins" = "ADMIN"
        "developers"      = "EDITOR"
      }
      useRoleAttribute = false
    }
  })
}

resource "signoz_auth_domain" "oidc" {
  name = "oidc.example.com"

  config = jsonencode({
    ssoEnabled = true
    ssoType    = "oidc"
    oidcConfig = {
      issuer                    = "https://login.example.com/realms/myrealm"
      issuerAlias               = "" # optional; when discovery issuer differs (e.g. Azure)
      clientId                  = "my-oidc-client-id"
      clientSecret              = "my-oidc-client-secret"
      getUserInfo               = true
      insecureSkipEmailVerified = false
      claimMapping = {
        email  = "email"
        name   = "name"
        groups = "groups"
        role   = "roles"
      }
    }
    roleMapping = {
      defaultRole = "VIEWER"
      groupMappings = {
        "platform-admins" = "ADMIN"
        "developers"      = "EDITOR"
      }
      useRoleAttribute = false
    }
  })
}

resource "signoz_auth_domain" "google" {
  name = "google.example.com"

  config = jsonencode({
    ssoEnabled = true
    ssoType    = "google_auth"
    googleAuthConfig = {
      clientId                       = "123456789.apps.googleusercontent.com"
      clientSecret                   = "my-google-client-secret"
      redirectURI                    = "https://signoz.example.com/api/v1/callback/google"
      fetchGroups                    = true
      fetchTransitiveGroupMembership = false
      insecureSkipEmailVerified      = false
      serviceAccountJson             = "{\"type\":\"service_account\",\"project_id\":\"my-project\",\"private_key_id\":\"key-id\",\"private_key\":\"-----BEGIN PRIVATE KEY-----\\n...\\n-----END PRIVATE KEY-----\\n\",\"client_email\":\"signoz@my-project.iam.gserviceaccount.com\",\"client_id\":\"123456789\",\"auth_uri\":\"https://accounts.google.com/o/oauth2/auth\",\"token_uri\":\"https://oauth2.googleapis.com/token\"}"
      domainToAdminEmail = {
        "example.com" = "admin@example.com"
      }
    }
    roleMapping = {
      defaultRole = "VIEWER"
      groupMappings = {
        "platform-admins" = "ADMIN"
        "developers"      = "EDITOR"
      }
      useRoleAttribute = false
    }
  })
}

resource "signoz_auth_domain" "password_only" {
  name = "password.example.com"

  config = jsonencode({
    ssoEnabled = false
    ssoType    = "email_password"
  })
}

# OIDC with role claim mapping instead of group mappings:
resource "signoz_auth_domain" "oidc_role_claim" {
  name = "oidc-role-claim.example.com"

  config = jsonencode({
    ssoEnabled = true
    ssoType    = "oidc"
    oidcConfig = {
      issuer                    = "https://login.example.com/realms/myrealm"
      clientId                  = "my-oidc-client-id"
      clientSecret              = "my-oidc-client-secret"
      getUserInfo               = true
      insecureSkipEmailVerified = false
      claimMapping = {
        email = "email"
        name  = "name"
        role  = "roles"
      }
    }
    roleMapping = {
      defaultRole      = "VIEWER"
      useRoleAttribute = true
      groupMappings    = {}
    }
  })
}
