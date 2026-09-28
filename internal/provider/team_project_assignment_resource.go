package provider

import (
	"context"
	"fmt"
	"terraform-provider-mixpanel/internal/mixpanel"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &teamProjectAssignmentResource{}
	_ resource.ResourceWithConfigure   = &teamProjectAssignmentResource{}
	_ resource.ResourceWithImportState = &teamProjectAssignmentResource{}
)

func NewTeamProjectAssignmentResource() resource.Resource {
	return &teamProjectAssignmentResource{}
}

type teamProjectAssignmentResource struct {
	client *mixpanel.Client
}

type TeamProjectAssignmentModel struct {
	Id        types.String `tfsdk:"id"`
	TeamId    types.Int64  `tfsdk:"team_id"`
	ProjectId types.Int64  `tfsdk:"project_id"`
	Role      types.String `tfsdk:"role"`
}

func (r *teamProjectAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mixpanel.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *mixpanel.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *teamProjectAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_project_assignment"
}

func (r *teamProjectAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gives the members of a team access to a project. Teams have no public API, this uses the endpoints of the Mixpanel UI.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "`<team_id>/<project_id>`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"team_id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Project role: `owner`, `admin`, `analyst` or `consumer`.",
				Required:            true,
			},
		},
	}
}

// readAssignment returns nil if the team doesn't exist or doesn't have the project.
func (r *teamProjectAssignmentResource) readAssignment(teamId, projectId int64) (*mixpanel.TeamProject, error) {
	team, err := r.client.GetTeam(teamId)
	if err != nil || team == nil {
		return nil, err
	}
	for _, project := range team.Projects {
		if project.Id == projectId {
			return &project, nil
		}
	}
	return nil, nil
}

// save adds the project to the team (or sets its role), then checks the API applied it.
func (r *teamProjectAssignmentResource) save(plan *TeamProjectAssignmentModel) diag.Diagnostics {
	var diags diag.Diagnostics
	teamId, projectId := plan.TeamId.ValueInt64(), plan.ProjectId.ValueInt64()

	err := r.client.AddProjectToTeam(teamId, projectId, plan.Role.ValueString())
	if err != nil {
		diags.AddError("Unable to add Mixpanel Project to Team", err.Error())
		return diags
	}

	project, err := r.readAssignment(teamId, projectId)
	if err != nil {
		diags.AddError("Unable to read Mixpanel Team Project Assignment", err.Error())
		return diags
	}
	if project == nil {
		diags.AddError(
			"Mixpanel Team Project Assignment not applied",
			fmt.Sprintf("Project %d is not assigned to team %d after the request.", projectId, teamId),
		)
		return diags
	}
	if project.Role != plan.Role.ValueString() {
		diags.AddError(
			"Mixpanel Team Project Assignment not applied",
			fmt.Sprintf("Team %d has the role %q on project %d after the request, instead of %q. If you changed the role, recreate the resource instead.", teamId, project.Role, projectId, plan.Role.ValueString()),
		)
		return diags
	}

	plan.Id = types.StringValue(fmt.Sprintf("%d/%d", teamId, projectId))
	return diags
}

func (r *teamProjectAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TeamProjectAssignmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.save(&plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Update changes the role in place by adding the project to the team again.
func (r *teamProjectAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TeamProjectAssignmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.save(&plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *teamProjectAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TeamProjectAssignmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.readAssignment(state.TeamId.ValueInt64(), state.ProjectId.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Mixpanel Team Project Assignment", err.Error())
		return
	}
	if project == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Id = types.StringValue(fmt.Sprintf("%d/%d", state.TeamId.ValueInt64(), state.ProjectId.ValueInt64()))
	state.Role = types.StringValue(project.Role)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *teamProjectAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TeamProjectAssignmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RemoveProjectFromTeam(state.TeamId.ValueInt64(), state.ProjectId.ValueInt64())
	if err != nil && !mixpanel.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to remove Mixpanel Project from Team", err.Error())
	}
}

func (r *teamProjectAssignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	teamId, projectId, err := parseIdPair(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", "Expected <team_id>/<project_id>: "+err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_id"), teamId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), projectId)...)
}
