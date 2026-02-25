package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sidequest-labs/terraform-provider-ancla/internal/client"
)

var _ datasource.DataSource = &AppDataSource{}

// AppDataSource reads an Ancla application.
type AppDataSource struct {
	client *client.Client
}

// AppDataSourceModel maps the data source schema data.
type AppDataSourceModel struct {
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

func NewAppDataSource() datasource.DataSource {
	return &AppDataSource{}
}

func (d *AppDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (d *AppDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an Ancla application by organization, project, and app slug.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the application.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The display name of the application.",
				Computed:    true,
			},
			"slug": schema.StringAttribute{
				Description: "The URL-friendly slug of the application.",
				Required:    true,
			},
			"organization_slug": schema.StringAttribute{
				Description: "The slug of the organization.",
				Required:    true,
			},
			"project_slug": schema.StringAttribute{
				Description: "The slug of the project.",
				Required:    true,
			},
			"platform": schema.StringAttribute{
				Description: "The platform type of the application.",
				Computed:    true,
			},
			"github_repository": schema.StringAttribute{
				Description: "The GitHub repository linked to this application.",
				Computed:    true,
			},
			"auto_deploy_branch": schema.StringAttribute{
				Description: "The branch that triggers automatic deployments.",
				Computed:    true,
			},
			"process_counts": schema.MapAttribute{
				Description: "Map of process type to replica count.",
				Computed:    true,
				ElementType: types.Int64Type,
			},
		},
	}
}

func (d *AppDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AppDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config AppDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := d.client.GetApp(
		config.OrganizationSlug.ValueString(),
		config.ProjectSlug.ValueString(),
		config.Slug.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error reading application", err.Error())
		return
	}

	config.ID = types.StringValue(app.ID)
	config.Name = types.StringValue(app.Name)
	config.Slug = types.StringValue(app.Slug)
	config.Platform = types.StringValue(app.Platform)
	config.GithubRepository = types.StringValue(app.GithubRepository)
	config.AutoDeployBranch = types.StringValue(app.AutoDeployBranch)

	if len(app.ProcessCounts) > 0 {
		elems := make(map[string]types.Int64, len(app.ProcessCounts))
		for k, v := range app.ProcessCounts {
			elems[k] = types.Int64Value(int64(v))
		}
		mapVal, diags := types.MapValueFrom(ctx, types.Int64Type, elems)
		resp.Diagnostics.Append(diags...)
		config.ProcessCounts = mapVal
	} else {
		config.ProcessCounts = types.MapNull(types.Int64Type)
	}

	diags = resp.State.Set(ctx, config)
	resp.Diagnostics.Append(diags...)
}
