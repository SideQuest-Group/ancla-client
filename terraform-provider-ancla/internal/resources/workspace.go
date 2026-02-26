package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sidequest-labs/terraform-provider-ancla/internal/client"
)

var (
	_ resource.Resource                = &WorkspaceResource{}
	_ resource.ResourceWithImportState = &WorkspaceResource{}
)

// WorkspaceResource manages an Ancla workspace.
type WorkspaceResource struct {
	client *client.Client
}

// WorkspaceResourceModel maps the resource schema data.
type WorkspaceResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Slug         types.String `tfsdk:"slug"`
	MemberCount  types.Int64  `tfsdk:"member_count"`
	ProjectCount types.Int64  `tfsdk:"project_count"`
}

func NewWorkspaceResource() resource.Resource {
	return &WorkspaceResource{}
}

func (r *WorkspaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (r *WorkspaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Ancla workspace.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the workspace.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The display name of the workspace.",
				Required:    true,
			},
			"slug": schema.StringAttribute{
				Description: "The URL-friendly slug of the workspace.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"member_count": schema.Int64Attribute{
				Description: "The number of members in the workspace.",
				Computed:    true,
			},
			"project_count": schema.Int64Attribute{
				Description: "The number of projects in the workspace.",
				Computed:    true,
			},
		},
	}
}

func (r *WorkspaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WorkspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkspaceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ws, err := r.client.CreateWorkspace(plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating workspace", err.Error())
		return
	}

	plan.ID = types.StringValue(ws.ID)
	plan.Slug = types.StringValue(ws.Slug)
	plan.Name = types.StringValue(ws.Name)
	plan.MemberCount = types.Int64Value(int64(ws.MemberCount))
	plan.ProjectCount = types.Int64Value(int64(ws.ProjectCount))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *WorkspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WorkspaceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ws, err := r.client.GetWorkspace(state.Slug.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading workspace", err.Error())
		return
	}

	state.ID = types.StringValue(ws.ID)
	state.Name = types.StringValue(ws.Name)
	state.Slug = types.StringValue(ws.Slug)
	state.MemberCount = types.Int64Value(int64(ws.MemberCount))
	state.ProjectCount = types.Int64Value(int64(ws.ProjectCount))

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *WorkspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkspaceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state WorkspaceResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ws, err := r.client.UpdateWorkspace(state.Slug.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating workspace", err.Error())
		return
	}

	plan.ID = types.StringValue(ws.ID)
	plan.Slug = types.StringValue(ws.Slug)
	plan.Name = types.StringValue(ws.Name)
	plan.MemberCount = types.Int64Value(int64(ws.MemberCount))
	plan.ProjectCount = types.Int64Value(int64(ws.ProjectCount))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *WorkspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WorkspaceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteWorkspace(state.Slug.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting workspace", err.Error())
		return
	}
}

func (r *WorkspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("slug"), req, resp)
}
