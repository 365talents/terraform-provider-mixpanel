resource "mixpanel_team_project_assignment" "developers" {
  team_id    = mixpanel_team.developers.id
  project_id = mixpanel_project.myproject.id
  role       = "analyst"
}
