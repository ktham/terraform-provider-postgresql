package provider

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"postgresql": providerserver.NewProtocol6WithError(New("test")()),
}

func providerConfig() string {
	dbUser := getEnv("DATABASE_USER", "terraform")
	dbPassword := getEnv("DATABASE_PASSWORD", "not_a_real_password")
	dbName := getEnv("DATABASE_NAME", "terraform_test")
	dbPort := getEnvAsInt("DATABASE_PORT", 15432)

	return fmt.Sprintf(`provider "postgresql" {
    hostname = "localhost"
    username = %q
    password = %q
    database_name = %q
    port = %d
  }`, dbUser, dbPassword, dbName, dbPort)
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		valInt, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return valInt
	}
	return defaultValue
}
