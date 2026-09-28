package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccTeamConfig(name, role string) string {
	return fmt.Sprintf(`
resource "mixpanel_team" "test" {
  name = %q
}

resource "mixpanel_team_project_assignment" "test" {
  team_id    = mixpanel_team.test.id
  project_id = %s
  role       = %q
}
`, name, os.Getenv("MIXPANEL_TEST_PROJECT_ID"), role)
}

// Creates a team and assigns it the persistent project MIXPANEL_TEST_PROJECT_ID.
// The team and the assignment are deleted at the end.
func TestAccTeam(t *testing.T) {
	name := "tf-acc-" + acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTeamConfig(name, "consumer"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mixpanel_team.test", "name", name),
					resource.TestCheckResourceAttrSet("mixpanel_team.test", "id"),
					resource.TestCheckResourceAttr("mixpanel_team_project_assignment.test", "role", "consumer"),
				),
			},
			{
				ResourceName:      "mixpanel_team.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "mixpanel_team_project_assignment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTeamConfig(name, "analyst"),
				Check:  resource.TestCheckResourceAttr("mixpanel_team_project_assignment.test", "role", "analyst"),
			},
			{
				Config: testAccTeamConfig(name+"-renamed", "analyst"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mixpanel_team.test", "name", name+"-renamed"),
					// Renamed in place, the assignment is kept.
					resource.TestCheckResourceAttr("mixpanel_team_project_assignment.test", "role", "analyst"),
				),
			},
		},
	})
}
