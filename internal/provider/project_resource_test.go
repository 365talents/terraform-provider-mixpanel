package provider

// Disabled: only owners can delete projects, and service accounts can be at
// most admin. The provider can't delete the project, so every run would leave
// one behind.
//
// import (
// 	"fmt"
// 	"testing"
//
// 	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
// 	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
// )
//
// func testAccProjectConfig(name string) string {
// 	return fmt.Sprintf(`
// resource "mixpanel_project" "test" {
//   name     = %q
//   domain   = "EU"
//   timezone = "Europe/Paris"
// }
//
// data "mixpanel_project" "test" {
//   id = mixpanel_project.test.id
// }
// `, name)
// }
//
// // Creates a project. It is not deleted at the end, see above.
// func TestAccProject(t *testing.T) {
// 	name := "tf-acc-" + acctest.RandString(8)
//
// 	resource.Test(t, resource.TestCase{
// 		PreCheck:                 func() { testAccPreCheck(t) },
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccProjectConfig(name),
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					resource.TestCheckResourceAttr("mixpanel_project.test", "name", name),
// 					resource.TestCheckResourceAttr("mixpanel_project.test", "domain", "EU"),
// 					resource.TestCheckResourceAttr("mixpanel_project.test", "timezone", "Europe/Paris"),
// 					resource.TestCheckResourceAttrSet("mixpanel_project.test", "token"),
// 					resource.TestCheckResourceAttrPair("data.mixpanel_project.test", "name", "mixpanel_project.test", "name"),
// 					resource.TestCheckResourceAttrPair("data.mixpanel_project.test", "token", "mixpanel_project.test", "token"),
// 				),
// 			},
// 			{
// 				ResourceName:      "mixpanel_project.test",
// 				ImportState:       true,
// 				ImportStateVerify: true,
// 			},
// 			{
// 				Config: testAccProjectConfig(name + "-renamed"),
// 				Check:  resource.TestCheckResourceAttr("mixpanel_project.test", "name", name+"-renamed"),
// 			},
// 		},
// 	})
// }
