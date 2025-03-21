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
			// Test role creation
			{
				Config: providerConfig() + testAccRoleResourceConfig("role1", true, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("postgresql_role.test", "name", "role1"),
					resource.TestCheckResourceAttr("postgresql_role.test", "can_login", "true"),
					resource.TestCheckResourceAttr("postgresql_role.test", "connection_limit", "10"),
				),
			},
			// Test role update
			{
				Config: providerConfig() + testAccRoleResourceConfig("role1", false, 15),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("postgresql_role.test", "name", "role1"),
					resource.TestCheckResourceAttr("postgresql_role.test", "can_login", "false"),
					resource.TestCheckResourceAttr("postgresql_role.test", "connection_limit", "15"),
				),
			},
			// Test role re-name
			{
				Config: providerConfig() + testAccRoleResourceConfig("role2", false, 15),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("postgresql_role.test", "name", "role2"),
					resource.TestCheckResourceAttr("postgresql_role.test", "can_login", "false"),
					resource.TestCheckResourceAttr("postgresql_role.test", "connection_limit", "15"),
				),
			},
		},
	})
}

func testAccRoleResourceConfig(name string, canLogin bool, connectionLimit int32) string {
	return fmt.Sprintf(`
resource "postgresql_role" "test" {
  name                      = %[1]q
  bypass_row_level_security = true
  can_login                 = %t
  connection_limit          = %d
  create_role               = true
  inherit                   = true
  replication               = true
  superuser                 = true
}
`, name, canLogin, connectionLimit)
}
