package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sidequest-labs/terraform-provider-ancla/internal/client"
)

var (
	_ resource.Resource                = &ConfigResource{}
	_ resource.ResourceWithImportState = &ConfigResource{}
)

// ConfigResource manages an Ancla configuration variable.
type ConfigResource struct {
	client *client.Client
}

// ConfigResourceModel maps the resource schema data.
type ConfigResourceModel struct {
	ID            types.String `tfsdk:"id"`
	WorkspaceSlug types.String `tfsdk:"workspace_slug"`
	ProjectSlug   types.String `tfsdk:"project_slug"`
	EnvSlug       types.String `tfsdk:"env_slug"`
	ServiceSlug   types.String `tfsdk:"service_slug"`
	Name          types.String `tfsdk:"name"`
	Value         types.String `tfsdk:"value"`
	Secret        types.Bool   `tfsdk:"secret"`
	Buildtime     types.Bool   `tfsdk:"buildtime"`
	Scope         types.String `tfsdk:"scope"`
}

func NewConfigResource() resource.Resource {
	return &ConfigResource{}
}

func (r *ConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_var"
}

func (r *ConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a configuration variable for an Ancla resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the configuration variable.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workspace_slug": schema.StringAttribute{
				Description: "The slug of the workspace.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_slug": schema.StringAttribute{
				Description: "The slug of the project. Required for project, environment, and service scopes.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"env_slug": schema.StringAttribute{
				Description: "The slug of the environment. Required for environment and service scopes.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_slug": schema.StringAttribute{
				Description: "The slug of the service. Required for service scope.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name (key) of the configuration variable.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Description: "The value of the configuration variable.",
				Required:    true,
				Sensitive:   true,
			},
			"secret": schema.BoolAttribute{
				Description: "Whether this variable is a secret (value hidden by default).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"buildtime": schema.BoolAttribute{
				Description: "Whether this variable is available at build time.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"scope": schema.StringAttribute{
				Description: "The scope of the configuration variable. One of: workspace, project, environment, service. Defaults to service.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("service"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *ConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigResource) configSlugs(model *ConfigResourceModel) (ws, proj, env, svc, scope string) {
	ws = model.WorkspaceSlug.ValueString()
	proj = model.ProjectSlug.ValueString()
	env = model.EnvSlug.ValueString()
	svc = model.ServiceSlug.ValueString()
	scope = model.Scope.ValueString()
	if scope == "" {
		scope = "service"
	}
	return
}

func (r *ConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ConfigResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ws, proj, env, svc, scope := r.configSlugs(&plan)

	cfg, err := r.client.SetConfig(
		ws, proj, env, svc, scope,
		plan.Name.ValueString(),
		plan.Value.ValueString(),
		plan.Secret.ValueBool(),
		plan.Buildtime.ValueBool(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating config variable", err.Error())
		return
	}

	plan.ID = types.StringValue(cfg.ID)
	plan.Name = types.StringValue(cfg.Name)
	plan.Value = types.StringValue(cfg.Value)
	plan.Secret = types.BoolValue(cfg.Secret)
	plan.Buildtime = types.BoolValue(cfg.Buildtime)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ConfigResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ws, proj, env, svc, scope := r.configSlugs(&state)

	configs, err := r.client.ListConfig(ws, proj, env, svc, scope)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading config variables", err.Error())
		return
	}

	// Find the config var by ID.
	var found *client.ConfigVar
	for i := range configs {
		if configs[i].ID == state.ID.ValueString() {
			found = &configs[i]
			break
		}
	}

	if found == nil {
		// Config var was deleted outside Terraform.
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(found.Name)
	state.Secret = types.BoolValue(found.Secret)
	state.Buildtime = types.BoolValue(found.Buildtime)
	// Only update value if it is not a secret (secrets come back masked).
	if !found.Secret {
		state.Value = types.StringValue(found.Value)
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ConfigResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ws, proj, env, svc, scope := r.configSlugs(&plan)

	// The API uses POST to upsert by name, so we POST again.
	cfg, err := r.client.SetConfig(
		ws, proj, env, svc, scope,
		plan.Name.ValueString(),
		plan.Value.ValueString(),
		plan.Secret.ValueBool(),
		plan.Buildtime.ValueBool(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating config variable", err.Error())
		return
	}

	plan.ID = types.StringValue(cfg.ID)
	plan.Name = types.StringValue(cfg.Name)
	plan.Value = types.StringValue(cfg.Value)
	plan.Secret = types.BoolValue(cfg.Secret)
	plan.Buildtime = types.BoolValue(cfg.Buildtime)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ConfigResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ws, proj, env, svc, scope := r.configSlugs(&state)

	if err := r.client.DeleteConfig(ws, proj, env, svc, scope, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting config variable", err.Error())
		return
	}
}

func (r *ConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: ws-slug/proj-slug/env-slug/svc-slug/config-id
	// For non-service scopes, use "-" as placeholder for unused segments.
	parts := strings.SplitN(req.ID, "/", 5)
	if len(parts) != 5 || parts[0] == "" || parts[4] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			"Expected import ID format: <workspace_slug>/<project_slug>/<env_slug>/<service_slug>/<config_id>. Use '-' for unused scope segments.")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_slug"), parts[0])...)
	if parts[1] != "-" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_slug"), parts[1])...)
	}
	if parts[2] != "-" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("env_slug"), parts[2])...)
	}
	if parts[3] != "-" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_slug"), parts[3])...)
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[4])...)
}
