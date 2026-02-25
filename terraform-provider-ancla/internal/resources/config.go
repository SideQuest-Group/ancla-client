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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sidequest-labs/terraform-provider-ancla/internal/client"
)

var (
	_ resource.Resource                = &ConfigResource{}
	_ resource.ResourceWithImportState = &ConfigResource{}
)

// ConfigResource manages an Ancla application configuration variable.
type ConfigResource struct {
	client *client.Client
}

// ConfigResourceModel maps the resource schema data.
type ConfigResourceModel struct {
	ID        types.String `tfsdk:"id"`
	AppID     types.String `tfsdk:"app_id"`
	Name      types.String `tfsdk:"name"`
	Value     types.String `tfsdk:"value"`
	Secret    types.Bool   `tfsdk:"secret"`
	Buildtime types.Bool   `tfsdk:"buildtime"`
}

func NewConfigResource() resource.Resource {
	return &ConfigResource{}
}

func (r *ConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config"
}

func (r *ConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a configuration variable for an Ancla application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the configuration variable.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_id": schema.StringAttribute{
				Description: "The application ID this configuration variable belongs to.",
				Required:    true,
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

func (r *ConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ConfigResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.SetConfig(
		plan.AppID.ValueString(),
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

	configs, err := r.client.ListConfig(state.AppID.ValueString())
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

	// The API uses POST to upsert by name, so we just POST again.
	cfg, err := r.client.SetConfig(
		plan.AppID.ValueString(),
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

	if err := r.client.DeleteConfig(state.AppID.ValueString(), state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting config variable", err.Error())
		return
	}
}

func (r *ConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: app-id/config-id
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			"Expected import ID format: <app_id>/<config_id>")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
