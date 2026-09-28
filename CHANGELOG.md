## 1.3.0 (2026-09-28)

FEATURES:

- New resource `mixpanel_service_account`. The secret is only available when the resource is created by Terraform, not after an import.
- New resource `mixpanel_service_account_project_membership`, to give a service account a role in a project.
- New resource `mixpanel_team`, to manage organization teams.
- New resource `mixpanel_team_project_assignment`, to give the members of a team a role in a project.

NOTES:

- Teams have no public Mixpanel API: `mixpanel_team` and `mixpanel_team_project_assignment` use the endpoints of the Mixpanel UI, which can change without notice.
- `mixpanel_project`: the docs now say that destroying a project only removes it from the Terraform state. Only owners can delete projects, and service accounts can be at most admin, so delete it in the Mixpanel UI. The behavior itself is unchanged.
- Updated all Go dependencies (terraform-plugin-framework 1.19, terraform-plugin-go 0.31, ...). Building the provider now requires Go 1.26.
- Added acceptance tests that run against the real Mixpanel API.

## 1.2.1 (2024-09-16)

BUG FIXES:

- `mixpanel_project`: `secret` was always empty.

## 1.2.0 (2024-09-16)

FEATURES:

- `mixpanel_project` resource and data source: new `secret` attribute (sensitive).

## 1.1.0 (2024-08-01)

FEATURES:

- Provider: new `concurrent_requests` option to limit the number of concurrent requests to Mixpanel (default: 3).
- Provider: failed requests to the Mixpanel API are retried with backoff.

## 1.0.0 (2024-07-31)

FEATURES:

- Provider authentication with a Mixpanel service account (`service_account_username` and `service_account_secret`, or the `MIXPANEL_SERVICE_ACCOUNT_USERNAME` and `MIXPANEL_SERVICE_ACCOUNT_SECRET` environment variables). The provider uses the first organization of the service account.
- New resource `mixpanel_project` (`name`, `domain`, `timezone`, and computed `api_key` and `token`), with import support.
- New data source `mixpanel_project`, to look up a project by name.
