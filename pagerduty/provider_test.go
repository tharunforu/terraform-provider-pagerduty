package pagerduty

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)
var testAccProviders map[string]*schema.Provider

var testAccProvider *schema.Provider
func init() {

	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"pagerduty": testAccProvider,
	}
}
func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("Token"); v == "" {
		t.Fatal("no token is assigned")
	}
}
