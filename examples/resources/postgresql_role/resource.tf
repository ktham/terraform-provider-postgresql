resource "postgresql_role" "example" {
  name             = "example"
  can_login        = true
  connection_limit = 25
  superuser        = false
}
