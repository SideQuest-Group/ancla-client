package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sidequest-labs/terraform-provider-ancla/internal/client"
)

var (
	_ resource.Resource                = &ServiceResource{}
	_ resource.ResourceWithImportState = &ServiceResource{}
)

// ServiceResource manages an Ancla service.
type ServiceResource struct {
	client *client.Client
}

// ServiceResourceModel maps the resource schema data.
type ServiceResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Slug             types.String `tfsdk:"slug"`
	WorkspaceSlug    types.String `tfsdk:"workspace_slug"`
	ProjectSlug      types.String `tfsdk:"project_slug"`
	EnvSlug          types.String `tfsdk:"env_slug"`
	Platform         types.String `tfsdk:"platform"`
	GithubRepository types.String `tfsdk:"github_repository"`
	AutoDeployBranch types.String `tfsdk:"auto_deploy_branch"`
	ProcessCounts    types.Map    `tfsdk:"process_counts"`
}

func NewServiceResource() resource.Resource {
	return &ServiceResource{}
}

func (r *ServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

func (r *ServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Ancla service within an environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the service.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The display name of the service.",
				Required:    true,
			},
			"slug": schema.StringAttribute{
				Description: "The URL-friendly slug of the service.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workspace_slug": schema.StringAttribute{
				Description: "The slug of the workspace this service belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_slug": schema.StringAttribute{
				Description: "The slug of the project this service belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"env_slug": schema.StringAttribute{
				Description: "The slug of the environment this service belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"platform": schema.StringAttribute{
				Description: "The platform type of the service.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"github_repository": schema.StringAttribute{
				Description: "The GitHub repository linked to this service.",
				Optional:    true,
				Computed:    true,
			},
			"auto_deploy_branch": schema.StringAttribute{
				Description: "The branch that triggers automatic deployments.",
				Optional:    true,
				Computed:    true,
			},
			"process_counts": schema.MapAttribute{
				Description: "Map of process type to replica count (e.g. web=2, worker=1).",
				Optional:    true,
				Computed:    true,
				ElementType: types.Int64Type,
			},
		},
	}
}

func (r *ServiceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *ServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ServiceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	svc, err := r.client.CreateService(
		plan.WorkspaceSlug.ValueString(),
		plan.ProjectSlug.ValueString(),
		plan.EnvSlug.ValueString(),
		plan.Name.ValueString(),
		plan.Platform.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating service", err.Error())
		return
	}

	r.mapServiceToState(ctx, svc, &plan, &resp.Diagnostics)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ServiceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	svc, err := r.client.GetService(
		state.WorkspaceSlug.ValueString(),
		state.ProjectSlug.ValueString(),
		state.EnvSlug.ValueString(),
		state.Slug.ValueString(),
	)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading service", err.Error())
		return
	}

	r.mapServiceToState(ctx, svc, &state, &resp.Diagnostics)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ServiceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ServiceResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fields := map[string]any{
		"name": plan.Name.ValueString(),
	}
	if !plan.GithubRepository.IsNull() && !plan.GithubRepository.IsUnknown() {
		fields["github_repository"] = plan.GithubRepository.ValueString()
	}
	if !plan.AutoDeployBranch.IsNull() && !plan.AutoDeployBranch.IsUnknown() {
		fields["auto_deploy_branch"] = plan.AutoDeployBranch.ValueString()
	}

	svc, err := r.client.UpdateService(
		state.WorkspaceSlug.ValueString(),
		state.ProjectSlug.ValueString(),
		state.EnvSlug.ValueString(),
		state.Slug.ValueString(),
		fields,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating service", err.Error())
		return
	}

	r.mapServiceToState(ctx, svc, &plan, &resp.Diagnostics)

	// Handle scale if process_counts changed.
	if !plan.ProcessCounts.IsNull() && !plan.ProcessCounts.IsUnknown() {
		var counts map[string]int64
		diags = plan.ProcessCounts.ElementsAs(ctx, &counts, false)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() && len(counts) > 0 {
			intCounts := make(map[string]int)
			for k, v := range counts {
				intCounts[k] = int(v)
			}
			if err := r.client.ScaleService(
				plan.WorkspaceSlug.ValueString(),
				plan.ProjectSlug.ValueString(),
				plan.EnvSlug.ValueString(),
				svc.Slug,
				intCounts,
			); err != nil {
				resp.Diagnostics.AddError("Error scaling service", err.Error())
				return
			}
			// Re-read to get updated process counts.
			svc, err = r.client.GetService(
				plan.WorkspaceSlug.ValueString(),
				plan.ProjectSlug.ValueString(),
				plan.EnvSlug.ValueString(),
				svc.Slug,
			)
			if err != nil {
				resp.Diagnostics.AddError("Error reading service after scale", err.Error())
				return
			}
			r.mapServiceToState(ctx, svc, &plan, &resp.Diagnostics)
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ServiceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteService(
		state.WorkspaceSlug.ValueString(),
		state.ProjectSlug.ValueString(),
		state.EnvSlug.ValueString(),
		state.Slug.ValueString(),
	); err != nil {
		resp.Diagnostics.AddError("Error deleting service", err.Error())
		return
	}
}

func (r *ServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: ws/proj/env/svc
	parts := strings.SplitN(req.ID, "/", 4)
	if len(parts) != 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			"Expected import ID format: <workspace_slug>/<project_slug>/<env_slug>/<service_slug>")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_slug"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_slug"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("env_slug"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("slug"), parts[3])...)
}

func (r *ServiceResource) mapServiceToState(ctx context.Context, svc *client.Service, model *ServiceResourceModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(svc.ID)
	model.Name = types.StringValue(svc.Name)
	model.Slug = types.StringValue(svc.Slug)
	model.Platform = types.StringValue(svc.Platform)

	if svc.GithubRepository != "" {
		model.GithubRepository = types.StringValue(svc.GithubRepository)
	} else if model.GithubRepository.IsNull() {
		model.GithubRepository = types.StringValue("")
	}

	if svc.AutoDeployBranch != "" {
		model.AutoDeployBranch = types.StringValue(svc.AutoDeployBranch)
	} else if model.AutoDeployBranch.IsNull() {
		model.AutoDeployBranch = types.StringValue("")
	}

	if len(svc.ProcessCounts) > 0 {
		elems := make(map[string]types.Int64, len(svc.ProcessCounts))
		for k, v := range svc.ProcessCounts {
			elems[k] = types.Int64Value(int64(v))
		}
		mapVal, d := types.MapValueFrom(ctx, types.Int64Type, elems)
		diags.Append(d...)
		model.ProcessCounts = mapVal
	} else {
		model.ProcessCounts = types.MapNull(types.Int64Type)
	}
}
