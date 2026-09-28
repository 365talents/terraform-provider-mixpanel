// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"mixpanel": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck fails the test when a Mixpanel env var is missing, so a
// missing secret in CI doesn't go unnoticed.
func testAccPreCheck(t *testing.T) {
	for _, name := range []string{"MIXPANEL_SERVICE_ACCOUNT_USERNAME", "MIXPANEL_SERVICE_ACCOUNT_SECRET", "MIXPANEL_TEST_PROJECT_ID"} {
		if os.Getenv(name) == "" {
			t.Fatalf("%s must be set for acceptance tests", name)
		}
	}
}
