package provider

import (
	"context"
	"fmt"
	"terraform-provider-mixpanel/internal/mixpanel"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &ProjectDataSource{}
	_ datasource.DataSourceWithConfigure = &ProjectDataSource{}
)

// NewprojectDataSource is a helper function to simplify the provider implementation.
func NewprojectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

// ProjectDataSource is the data source implementation.
type ProjectDataSource struct {
	client *mixpanel.Client
}

type ProjectDataSourceModel = ProjectModel

// Metadata returns the data source type name.
func (d *ProjectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema defines the schema for the data source.
func (d *ProjectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"domain": schema.StringAttribute{
				Computed: true,
			},
			"timezone": schema.StringAttribute{
				Computed: true,
			},
			"api_key": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"token": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"secret": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

type ProjectModel struct {
	Id       types.Int64           `tfsdk:"id"`
	Name     basetypes.StringValue `tfsdk:"name"`
	Domain   basetypes.StringValue `tfsdk:"domain"`
	Timezone basetypes.StringValue `tfsdk:"timezone"`
	ApiKey   basetypes.StringValue `tfsdk:"api_key"`
	Token    basetypes.StringValue `tfsdk:"token"`
	Secret   basetypes.StringValue `tfsdk:"secret"`
}

// Read refreshes the Terraform state with the latest data.
func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var id int64

	idPath := path.Root("id")
	// Retrieve the id from the terraform data source
	req.Config.GetAttribute(ctx, idPath, &id)

	project, err := d.client.GetProject(id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Mixpanel Project",
			err.Error(),
		)
		return
	}

	projectState := ProjectToProjectModel(project)

	// Set state
	diags := resp.State.Set(ctx, &projectState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProjectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*mixpanel.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *mixpanel.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}
