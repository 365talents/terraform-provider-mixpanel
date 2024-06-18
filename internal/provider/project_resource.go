package provider

import (
	"context"
	"fmt"
	"strconv"
	"terraform-provider-mixpanel/internal/mixpanel"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &projectResource{}
	_ resource.ResourceWithConfigure   = &projectResource{}
	_ resource.ResourceWithImportState = &projectResource{}
)

// NewProjectResource is a helper function to simplify the provider implementation.
func NewProjectResource() resource.Resource {
	return &projectResource{}
}

// projectResource is the resource implementation.
type projectResource struct {
	client *mixpanel.Client
}

// Configure adds the provider configured client to the resource.
func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

// Metadata returns the resource type name.
func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema defines the schema for the resource.
func (r *projectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"domain": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"timezone": schema.StringAttribute{
				Required: true,
			},
			"api_key": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"token": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProjectModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed order value from HashiCups
	project, err := r.client.GetProject(state.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Mixpanel Project",
			"Could not read Mixpanel project ID "+strconv.FormatInt(state.Id.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	// Update the state with the refreshed data
	state.Name = basetypes.NewStringValue(project.Name)
	state.Timezone = basetypes.NewStringValue(project.Timezone)

	state = ProjectToProjectModel(project)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Update updates the resource and sets the updated Terraform state on success.
func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProjectModel
	var state ProjectModel

	r.client.GetTimezones()

	// Get the planned new state
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Get the current state
	diags = req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if plan.Name != state.Name {
		err := r.client.UpdateProjectName(state.Id.ValueInt64(), plan.Name.String())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update Mixpanel Project Name",
				err.Error(),
			)
			return
		}
	}

	if plan.Timezone != state.Timezone {
		err := r.client.UpdateProjectTimezone(state.Id.ValueInt64(), plan.Timezone.String())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update Mixpanel Project Timezone",
				err.Error(),
			)
			return
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Not Implemented, Service Account does not have permission to delete projects and we don't support any other authentication method yet
}

// Create creates the resource and sets the initial Terraform state.
func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProjectModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data := mixpanel.Project{
		Name:     plan.Name.String(),
		Domain:   plan.Domain.String(),
		Timezone: plan.Timezone.String(),
	}

	newProject, err := r.client.CreateProject(&data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Mixpanel Project",
			err.Error(),
		)
		return
	}

	project, err := r.client.GetProject(newProject.Id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Mixpanel Project",
			err.Error(),
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, ProjectToProjectModel(project))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute

	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid ID",
			"ID must be an integer",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func ProjectToProjectModel(project *mixpanel.Project) ProjectModel {
	return ProjectModel{
		Id:       types.Int64Value(project.Id),
		Name:     basetypes.NewStringValue(project.Name),
		Domain:   basetypes.NewStringValue(project.Domain),
		Timezone: basetypes.NewStringValue(project.Timezone),
		ApiKey:   basetypes.NewStringValue(project.ApiKey),
		Token:    basetypes.NewStringValue(project.Token),
	}
}
