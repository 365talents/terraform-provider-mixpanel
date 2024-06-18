// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"terraform-provider-mixpanel/internal/mixpanel"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure MixpanelProvider satisfies various provider interfaces.
var _ provider.Provider = &MixpanelProvider{}
var _ provider.ProviderWithFunctions = &MixpanelProvider{}

// MixpanelProvider defines the provider implementation.
type MixpanelProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *MixpanelProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mixpanel"
	resp.Version = p.version
}

// MixpanelProviderModel maps provider schema data to a Go type.
type MixpanelProviderModel struct {
	ServiceAccountUsername types.String `tfsdk:"service_account_username"`
	ServiceAccountSecret types.String `tfsdk:"service_account_secret"`
}

func (p *MixpanelProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"service_account_username": schema.StringAttribute{
				MarkdownDescription: "Mixpanel Service Account username",
				Optional:            true,
			},
			"service_account_secret": schema.StringAttribute{
				MarkdownDescription: "Mixpanel Service Account secret",
				Optional:            true,
				Sensitive: 				 	 true,
			},
		},
	}
}

func (p *MixpanelProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config MixpanelProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.ServiceAccountUsername.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
				path.Root("service_account_username"),
				"Unknown Mixpanel Service Account Username",
				"The provider cannot create the Mixpanel API client as there is an unknown configuration value for the Mixpanel Service Account Username. "+
						"Either target apply the source of the value first, set the value statically in the configuration, or use the MIXPANEL_SERVICE_ACCOUNT_USERNAME environment variable.",
		)
	}
	if config.ServiceAccountSecret.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
					path.Root("service_account_secret"),
					"Unknown Mixpanel Service Account Secret",
					"The provider cannot create the Mixpanel API client as there is an unknown configuration value for the Mixpanel Service Account secret. "+
							"Either target apply the source of the value first, set the value statically in the configuration, or use the MIXPANEL_SERVICE_ACCOUNT_SECRET environment variable.",
			)
	}
	if resp.Diagnostics.HasError() {
			return
	}

	serviceAccountUsername := os.Getenv("MIXPANEL_SERVICE_ACCOUNT_USERNAME")
	serviceAccountSecret := os.Getenv("MIXPANEL_SERVICE_ACCOUNT_SECRET")

	if !config.ServiceAccountUsername.IsNull() {
		serviceAccountUsername = config.ServiceAccountUsername.ValueString()
	}

	if !config.ServiceAccountSecret.IsNull() {
		serviceAccountSecret = config.ServiceAccountSecret.ValueString()
	}

	if serviceAccountUsername == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("service_account_username"),
			"Mixpanel Service Account Username is required",
			"The provider cannot create the Mixpanel API client as there is a missing or empty configuration value for the Mixpanel Service Account Username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MIXPANEL_SERVICE_ACCOUNT_USERNAME environment variable.",
		)
	}

	if serviceAccountSecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("service_account_secret"),
			"Mixpanel Service Account Secret is required",
			"The provider cannot create the Mixpanel API client as there is missing or empty configuration value for the Mixpanel Service Account password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MIXPANEL_SERVICE_ACCOUNT_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}


	// Create the Mixpanel API client
	client, err := mixpanel.NewClient(&serviceAccountUsername, &serviceAccountSecret)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Mixpanel API client", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *MixpanelProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
		NewProjectResource,
	}
}

func (p *MixpanelProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleDataSource,
		NewprojectDataSource,
	}
}

func (p *MixpanelProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewExampleFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MixpanelProvider{
			version: version,
		}
	}
}
