package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccRoleResourceConfig("role1"),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO: Implement
				// resource.TestCheckResourceAttr("postgresql_role.test", "name", "role1"),
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + testAccRoleResourceConfig("role2"),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO: Implement
				// resource.TestCheckResourceAttr("postgresql_role.test", "name", "role2"),
				),
			},
		},
	})
}

func testAccRoleResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "postgresql_role" "test" {
  # TODO:
  # name = %[1]q
}
`, name)
}
