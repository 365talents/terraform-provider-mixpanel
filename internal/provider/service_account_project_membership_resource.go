package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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
	_ resource.Resource                = &serviceAccountProjectMembershipResource{}
	_ resource.ResourceWithConfigure   = &serviceAccountProjectMembershipResource{}
	_ resource.ResourceWithImportState = &serviceAccountProjectMembershipResource{}
)

func NewServiceAccountProjectMembershipResource() resource.Resource {
	return &serviceAccountProjectMembershipResource{}
}

type serviceAccountProjectMembershipResource struct {
	client *mixpanel.Client
}

type ServiceAccountProjectMembershipModel struct {
	Id               types.String `tfsdk:"id"`
	ServiceAccountId types.Int64  `tfsdk:"service_account_id"`
	ProjectId        types.Int64  `tfsdk:"project_id"`
	Role             types.String `tfsdk:"role"`
}

func (r *serviceAccountProjectMembershipResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serviceAccountProjectMembershipResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account_project_membership"
}

func (r *serviceAccountProjectMembershipResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Membership of a service account in a project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "`<service_account_id>/<project_id>`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_account_id": schema.Int64Attribute{
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
				MarkdownDescription: "Project role: `admin`, `analyst` or `consumer`. Service accounts can't be project owners.",
				Required:            true,
			},
		},
	}
}

// readMembership returns nil if the service account is not a member of the project.
func (r *serviceAccountProjectMembershipResource) readMembership(serviceAccountId, projectId int64) (*mixpanel.ServiceAccountProjectMember, error) {
	members, err := r.client.GetProjectServiceAccounts(projectId)
	if mixpanel.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	for _, member := range members {
		if member.Id == serviceAccountId {
			return &member, nil
		}
	}
	return nil, nil
}

// save adds the membership (or sets its role), then checks the API applied it.
func (r *serviceAccountProjectMembershipResource) save(plan *ServiceAccountProjectMembershipModel) diag.Diagnostics {
	var diags diag.Diagnostics
	serviceAccountId, projectId := plan.ServiceAccountId.ValueInt64(), plan.ProjectId.ValueInt64()

	err := r.client.AddServiceAccountToProject(serviceAccountId, projectId, plan.Role.ValueString())
	if err != nil {
		diags.AddError("Unable to add Mixpanel Service Account to Project", err.Error())
		return diags
	}

	member, err := r.readMembership(serviceAccountId, projectId)
	if err != nil {
		diags.AddError("Unable to read Mixpanel Service Account Project Membership", err.Error())
		return diags
	}
	if member == nil {
		diags.AddError(
			"Mixpanel Service Account Project Membership not applied",
			fmt.Sprintf("Service account %d is not a member of project %d after the request.", serviceAccountId, projectId),
		)
		return diags
	}
	if member.Role != plan.Role.ValueString() {
		diags.AddError(
			"Mixpanel Service Account Project Membership not applied",
			fmt.Sprintf("Service account %d has the role %q on project %d after the request, instead of %q. If you changed the role, recreate the resource instead.", serviceAccountId, member.Role, projectId, plan.Role.ValueString()),
		)
		return diags
	}

	plan.Id = types.StringValue(fmt.Sprintf("%d/%d", serviceAccountId, projectId))
	return diags
}

func (r *serviceAccountProjectMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ServiceAccountProjectMembershipModel
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

// Update changes the role in place by adding the membership again.
func (r *serviceAccountProjectMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ServiceAccountProjectMembershipModel
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

func (r *serviceAccountProjectMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ServiceAccountProjectMembershipModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	member, err := r.readMembership(state.ServiceAccountId.ValueInt64(), state.ProjectId.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Mixpanel Service Account Project Membership", err.Error())
		return
	}
	if member == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Id = types.StringValue(fmt.Sprintf("%d/%d", state.ServiceAccountId.ValueInt64(), state.ProjectId.ValueInt64()))
	state.Role = types.StringValue(member.Role)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *serviceAccountProjectMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ServiceAccountProjectMembershipModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RemoveServiceAccountFromProject(state.ServiceAccountId.ValueInt64(), state.ProjectId.ValueInt64())
	if err != nil && !mixpanel.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to remove Mixpanel Service Account from Project", err.Error())
	}
}

func (r *serviceAccountProjectMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	serviceAccountId, projectId, err := parseMembershipId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_account_id"), serviceAccountId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), projectId)...)
}

func parseMembershipId(id string) (int64, int64, error) {
	parts := strings.Split(id, "/")
	if len(parts) == 2 {
		serviceAccountId, err1 := strconv.ParseInt(parts[0], 10, 64)
		projectId, err2 := strconv.ParseInt(parts[1], 10, 64)
		if err1 == nil && err2 == nil {
			return serviceAccountId, projectId, nil
		}
	}
	return 0, 0, fmt.Errorf("ID must be <service_account_id>/<project_id>, got %q", id)
}
