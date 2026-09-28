package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccServiceAccountConfig(username, role string) string {
	return fmt.Sprintf(`
resource "mixpanel_service_account" "test" {
  username = %q
}

resource "mixpanel_service_account_project_membership" "test" {
  service_account_id = mixpanel_service_account.test.id
  project_id         = %s
  role               = %q
}
`, username, os.Getenv("MIXPANEL_TEST_PROJECT_ID"), role)
}

// Creates a service account and adds it to the persistent project MIXPANEL_TEST_PROJECT_ID.
// The service account and the membership are deleted at the end.
func TestAccServiceAccount(t *testing.T) {
	username := "tf-acc-" + acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountConfig(username, "consumer"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mixpanel_service_account.test", "username", username),
					resource.TestMatchResourceAttr("mixpanel_service_account.test", "full_username", regexp.MustCompile(`^`+username+`\.[^.]+\.mp-service-account$`)),
					resource.TestCheckResourceAttrSet("mixpanel_service_account.test", "id"),
					resource.TestCheckResourceAttrSet("mixpanel_service_account.test", "secret"),
					resource.TestCheckResourceAttr("mixpanel_service_account_project_membership.test", "role", "consumer"),
				),
			},
			{
				ResourceName:      "mixpanel_service_account.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Not returned by the API.
				ImportStateVerifyIgnore: []string{"secret"},
			},
			{
				ResourceName:      "mixpanel_service_account_project_membership.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceAccountConfig(username, "analyst"),
				Check:  resource.TestCheckResourceAttr("mixpanel_service_account_project_membership.test", "role", "analyst"),
			},
		},
	})
}
