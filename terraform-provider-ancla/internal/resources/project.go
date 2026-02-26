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
	_ resource.Resource                = &ProjectResource{}
	_ resource.ResourceWithImportState = &ProjectResource{}
)

// ProjectResource manages an Ancla project.
type ProjectResource struct {
	client *client.Client
}

// ProjectResourceModel maps the resource schema data.
type ProjectResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Slug          types.String `tfsdk:"slug"`
	WorkspaceSlug types.String `tfsdk:"workspace_slug"`
	ServiceCount  types.Int64  `tfsdk:"service_count"`
}

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

func (r *ProjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *ProjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Ancla project within a workspace.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the project.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The display name of the project.",
				Required:    true,
			},
			"slug": schema.StringAttribute{
				Description: "The URL-friendly slug of the project.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workspace_slug": schema.StringAttribute{
				Description: "The slug of the workspace this project belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_count": schema.Int64Attribute{
				Description: "The number of services in the project.",
				Computed:    true,
			},
		},
	}
}

func (r *ProjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProjectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.CreateProject(plan.WorkspaceSlug.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating project", err.Error())
		return
	}

	plan.ID = types.StringValue(project.ID)
	plan.Slug = types.StringValue(project.Slug)
	plan.Name = types.StringValue(project.Name)
	plan.WorkspaceSlug = types.StringValue(project.WorkspaceSlug)
	plan.ServiceCount = types.Int64Value(int64(project.ServiceCount))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProjectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.GetProject(state.WorkspaceSlug.ValueString(), state.Slug.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project", err.Error())
		return
	}

	state.ID = types.StringValue(project.ID)
	state.Name = types.StringValue(project.Name)
	state.Slug = types.StringValue(project.Slug)
	state.WorkspaceSlug = types.StringValue(project.WorkspaceSlug)
	state.ServiceCount = types.Int64Value(int64(project.ServiceCount))

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProjectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ProjectResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.UpdateProject(
		state.WorkspaceSlug.ValueString(),
		state.Slug.ValueString(),
		plan.Name.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
		return
	}

	plan.ID = types.StringValue(project.ID)
	plan.Slug = types.StringValue(project.Slug)
	plan.Name = types.StringValue(project.Name)
	plan.WorkspaceSlug = types.StringValue(project.WorkspaceSlug)
	plan.ServiceCount = types.Int64Value(int64(project.ServiceCount))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProjectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteProject(state.WorkspaceSlug.ValueString(), state.Slug.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting project", err.Error())
		return
	}
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: ws-slug/project-slug
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			"Expected import ID format: <workspace_slug>/<project_slug>")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_slug"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("slug"), parts[1])...)
}
