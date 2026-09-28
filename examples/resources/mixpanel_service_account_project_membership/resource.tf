resource "mixpanel_service_account_project_membership" "backend" {
  service_account_id = mixpanel_service_account.backend.id
  project_id         = mixpanel_project.myproject.id
  role               = "consumer"
}
