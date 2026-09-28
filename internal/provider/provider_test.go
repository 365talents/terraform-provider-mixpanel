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
// nolint:unused // Used by the acceptance tests, currently disabled.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"mixpanel": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck skips the test when the Mixpanel credentials are missing,
// so the suite stays green without secrets (e.g. on PRs from forks).
// nolint:unused // Used by the acceptance tests, currently disabled.
func testAccPreCheck(t *testing.T) {
	for _, name := range []string{"MIXPANEL_SERVICE_ACCOUNT_USERNAME", "MIXPANEL_SERVICE_ACCOUNT_SECRET"} {
		if os.Getenv(name) == "" {
			t.Skipf("%s must be set for acceptance tests", name)
		}
	}
}
