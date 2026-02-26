package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sidequest-labs/terraform-provider-ancla/internal/client"
)

var (
	_ resource.Resource                = &EnvironmentResource{}
	_ resource.ResourceWithImportState = &EnvironmentResource{}
)

// EnvironmentResource manages an Ancla environment.
type EnvironmentResource struct {
	client *client.Client
}

// EnvironmentResourceModel maps the resource schema data.
type EnvironmentResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Slug          types.String `tfsdk:"slug"`
	WorkspaceSlug types.String `tfsdk:"workspace_slug"`
	ProjectSlug   types.String `tfsdk:"project_slug"`
	ServiceCount  types.Int64  `tfsdk:"service_count"`
}

func NewEnvironmentResource() resource.Resource {
	return &EnvironmentResource{}
}

func (r *EnvironmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *EnvironmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Ancla environment within a project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the environment.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The display name of the environment.",
				Required:    true,
			},
			"slug": schema.StringAttribute{
				Description: "The URL-friendly slug of the environment.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workspace_slug": schema.StringAttribute{
				Description: "The slug of the workspace this environment belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_slug": schema.StringAttribute{
				Description: "The slug of the project this environment belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_count": schema.Int64Attribute{
				Description: "The number of services in the environment.",
				Computed:    true,
			},
		},
	}
}

func (r *EnvironmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EnvironmentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, err := r.client.CreateEnvironment(
		plan.WorkspaceSlug.ValueString(),
		plan.ProjectSlug.ValueString(),
		plan.Name.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating environment", err.Error())
		return
	}

	plan.ID = types.StringValue(env.ID)
	plan.Slug = types.StringValue(env.Slug)
	plan.Name = types.StringValue(env.Name)
	plan.ServiceCount = types.Int64Value(int64(env.ServiceCount))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *EnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EnvironmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, err := r.client.GetEnvironment(
		state.WorkspaceSlug.ValueString(),
		state.ProjectSlug.ValueString(),
		state.Slug.ValueString(),
	)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading environment", err.Error())
		return
	}

	state.ID = types.StringValue(env.ID)
	state.Name = types.StringValue(env.Name)
	state.Slug = types.StringValue(env.Slug)
	state.ServiceCount = types.Int64Value(int64(env.ServiceCount))

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *EnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EnvironmentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state EnvironmentResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, err := r.client.UpdateEnvironment(
		state.WorkspaceSlug.ValueString(),
		state.ProjectSlug.ValueString(),
		state.Slug.ValueString(),
		plan.Name.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating environment", err.Error())
		return
	}

	plan.ID = types.StringValue(env.ID)
	plan.Slug = types.StringValue(env.Slug)
	plan.Name = types.StringValue(env.Name)
	plan.ServiceCount = types.Int64Value(int64(env.ServiceCount))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *EnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EnvironmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteEnvironment(
		state.WorkspaceSlug.ValueString(),
		state.ProjectSlug.ValueString(),
		state.Slug.ValueString(),
	); err != nil {
		resp.Diagnostics.AddError("Error deleting environment", err.Error())
		return
	}
}

func (r *EnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: ws-slug/proj-slug/env-slug
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			"Expected import ID format: <workspace_slug>/<project_slug>/<env_slug>")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_slug"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_slug"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("slug"), parts[2])...)
}
