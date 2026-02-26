package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sidequest-labs/terraform-provider-ancla/internal/client"
)

var _ datasource.DataSource = &EnvironmentDataSource{}

// EnvironmentDataSource reads an Ancla environment.
type EnvironmentDataSource struct {
	client *client.Client
}

// EnvironmentDataSourceModel maps the data source schema data.
type EnvironmentDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Slug          types.String `tfsdk:"slug"`
	WorkspaceSlug types.String `tfsdk:"workspace_slug"`
	ProjectSlug   types.String `tfsdk:"project_slug"`
	ServiceCount  types.Int64  `tfsdk:"service_count"`
	Created       types.String `tfsdk:"created"`
}

func NewEnvironmentDataSource() datasource.DataSource {
	return &EnvironmentDataSource{}
}

func (d *EnvironmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (d *EnvironmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an Ancla environment by workspace, project, and environment slug.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the environment.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The display name of the environment.",
				Computed:    true,
			},
			"slug": schema.StringAttribute{
				Description: "The URL-friendly slug of the environment.",
				Required:    true,
			},
			"workspace_slug": schema.StringAttribute{
				Description: "The slug of the workspace.",
				Required:    true,
			},
			"project_slug": schema.StringAttribute{
				Description: "The slug of the project.",
				Required:    true,
			},
			"service_count": schema.Int64Attribute{
				Description: "The number of services in the environment.",
				Computed:    true,
			},
			"created": schema.StringAttribute{
				Description: "The creation timestamp of the environment.",
				Computed:    true,
			},
		},
	}
}

func (d *EnvironmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *EnvironmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config EnvironmentDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, err := d.client.GetEnvironment(
		config.WorkspaceSlug.ValueString(),
		config.ProjectSlug.ValueString(),
		config.Slug.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error reading environment", err.Error())
		return
	}

	config.ID = types.StringValue(env.ID)
	config.Name = types.StringValue(env.Name)
	config.Slug = types.StringValue(env.Slug)
	config.ServiceCount = types.Int64Value(int64(env.ServiceCount))
	config.Created = types.StringValue(env.Created)

	diags = resp.State.Set(ctx, config)
	resp.Diagnostics.Append(diags...)
}
