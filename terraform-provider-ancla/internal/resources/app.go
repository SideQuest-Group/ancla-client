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
	_ resource.Resource                = &AppResource{}
	_ resource.ResourceWithImportState = &AppResource{}
)

// AppResource manages an Ancla application.
type AppResource struct {
	client *client.Client
}

// AppResourceModel maps the resource schema data.
type AppResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Slug             types.String `tfsdk:"slug"`
	OrganizationSlug types.String `tfsdk:"organization_slug"`
	ProjectSlug      types.String `tfsdk:"project_slug"`
	Platform         types.String `tfsdk:"platform"`
	GithubRepository types.String `tfsdk:"github_repository"`
	AutoDeployBranch types.String `tfsdk:"auto_deploy_branch"`
	ProcessCounts    types.Map    `tfsdk:"process_counts"`
}

func NewAppResource() resource.Resource {
	return &AppResource{}
}

func (r *AppResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (r *AppResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Ancla application within a project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the application.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The display name of the application.",
				Required:    true,
			},
			"slug": schema.StringAttribute{
				Description: "The URL-friendly slug of the application.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_slug": schema.StringAttribute{
				Description: "The slug of the organization this application belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_slug": schema.StringAttribute{
				Description: "The slug of the project this application belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"platform": schema.StringAttribute{
				Description: "The platform type of the application.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"github_repository": schema.StringAttribute{
				Description: "The GitHub repository linked to this application.",
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

func (r *AppResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AppResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.CreateApp(
		plan.OrganizationSlug.ValueString(),
		plan.ProjectSlug.ValueString(),
		plan.Name.ValueString(),
		plan.Platform.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating application", err.Error())
		return
	}

	r.mapAppToState(ctx, app, &plan, &resp.Diagnostics)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *AppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AppResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.GetApp(
		state.OrganizationSlug.ValueString(),
		state.ProjectSlug.ValueString(),
		state.Slug.ValueString(),
	)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading application", err.Error())
		return
	}

	r.mapAppToState(ctx, app, &state, &resp.Diagnostics)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *AppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AppResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AppResourceModel
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

	app, err := r.client.UpdateApp(
		state.OrganizationSlug.ValueString(),
		state.ProjectSlug.ValueString(),
		state.Slug.ValueString(),
		fields,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating application", err.Error())
		return
	}

	r.mapAppToState(ctx, app, &plan, &resp.Diagnostics)

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
			if err := r.client.ScaleApp(app.ID, intCounts); err != nil {
				resp.Diagnostics.AddError("Error scaling application", err.Error())
				return
			}
			// Re-read to get updated process counts.
			app, err = r.client.GetApp(
				plan.OrganizationSlug.ValueString(),
				plan.ProjectSlug.ValueString(),
				app.Slug,
			)
			if err != nil {
				resp.Diagnostics.AddError("Error reading application after scale", err.Error())
				return
			}
			r.mapAppToState(ctx, app, &plan, &resp.Diagnostics)
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *AppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AppResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteApp(
		state.OrganizationSlug.ValueString(),
		state.ProjectSlug.ValueString(),
		state.Slug.ValueString(),
	); err != nil {
		resp.Diagnostics.AddError("Error deleting application", err.Error())
		return
	}
}

func (r *AppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: org-slug/project-slug/app-slug
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			"Expected import ID format: <organization_slug>/<project_slug>/<app_slug>")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_slug"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_slug"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("slug"), parts[2])...)
}

func (r *AppResource) mapAppToState(ctx context.Context, app *client.App, model *AppResourceModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(app.ID)
	model.Name = types.StringValue(app.Name)
	model.Slug = types.StringValue(app.Slug)
	model.Platform = types.StringValue(app.Platform)

	if app.GithubRepository != "" {
		model.GithubRepository = types.StringValue(app.GithubRepository)
	} else if model.GithubRepository.IsNull() {
		model.GithubRepository = types.StringValue("")
	}

	if app.AutoDeployBranch != "" {
		model.AutoDeployBranch = types.StringValue(app.AutoDeployBranch)
	} else if model.AutoDeployBranch.IsNull() {
		model.AutoDeployBranch = types.StringValue("")
	}

	if len(app.ProcessCounts) > 0 {
		elems := make(map[string]types.Int64, len(app.ProcessCounts))
		for k, v := range app.ProcessCounts {
			elems[k] = types.Int64Value(int64(v))
		}
		mapVal, d := types.MapValueFrom(ctx, types.Int64Type, elems)
		diags.Append(d...)
		model.ProcessCounts = mapVal
	} else {
		model.ProcessCounts = types.MapNull(types.Int64Type)
	}
}
