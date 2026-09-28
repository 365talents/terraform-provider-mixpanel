package provider

import (
	"fmt"
	"os"
	"strconv"
	"terraform-provider-mixpanel/internal/mixpanel"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccProjectConfig(project *mixpanel.Project, name string) string {
	return fmt.Sprintf(`
resource "mixpanel_project" "test" {
  name     = %q
  domain   = %q
  timezone = %q
}

data "mixpanel_project" "test" {
  id = mixpanel_project.test.id
}
`, name, project.Domain, project.Timezone)
}

// Service accounts can't delete projects, so the test imports the persistent
// project MIXPANEL_TEST_PROJECT_ID instead of creating one. Destroying a
// mixpanel_project only removes it from the state, the project is kept.
func TestAccProject(t *testing.T) {
	// The config is built from the current project, read it only when the test runs.
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for acceptance tests")
	}
	testAccPreCheck(t)

	username := os.Getenv("MIXPANEL_SERVICE_ACCOUNT_USERNAME")
	secret := os.Getenv("MIXPANEL_SERVICE_ACCOUNT_SECRET")
	client, err := mixpanel.NewClient(&username, &secret, 1)
	if err != nil {
		t.Fatal(err)
	}
	id, err := strconv.ParseInt(os.Getenv("MIXPANEL_TEST_PROJECT_ID"), 10, 64)
	if err != nil {
		t.Fatalf("MIXPANEL_TEST_PROJECT_ID must be an integer: %s", err)
	}
	project, err := client.GetProject(id)
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccProjectConfig(project, project.Name),
				ResourceName:       "mixpanel_project.test",
				ImportState:        true,
				ImportStateId:      strconv.FormatInt(id, 10),
				ImportStatePersist: true,
			},
			{
				// Also fails if the import leaves a diff.
				Config: testAccProjectConfig(project, project.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mixpanel_project.test", "name", project.Name),
					resource.TestCheckResourceAttrSet("mixpanel_project.test", "token"),
					resource.TestCheckResourceAttrPair("data.mixpanel_project.test", "name", "mixpanel_project.test", "name"),
					resource.TestCheckResourceAttrPair("data.mixpanel_project.test", "token", "mixpanel_project.test", "token"),
				),
			},
			{
				Config: testAccProjectConfig(project, project.Name+"-renamed"),
				Check:  resource.TestCheckResourceAttr("mixpanel_project.test", "name", project.Name+"-renamed"),
			},
			{
				// Restore the name, the destroy doesn't touch the project.
				Config: testAccProjectConfig(project, project.Name),
				Check:  resource.TestCheckResourceAttr("mixpanel_project.test", "name", project.Name),
			},
		},
	})
}
