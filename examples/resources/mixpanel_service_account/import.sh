# Service accounts can be imported by their numeric id.
# The secret is only returned by Mixpanel at creation, so it stays null after an import.
terraform import mixpanel_service_account.backend 123
