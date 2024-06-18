provider "mixpanel" {
  service_account_username = "foo.mp-service-account"
  service_account_secret = "" # Prefer using an environment variable for this
}