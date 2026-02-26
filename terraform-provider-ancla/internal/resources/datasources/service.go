package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sidequest-labs/terraform-provider-ancla/internal/client"
)

var _ datasource.DataSource = &ServiceDataSource{}

// ServiceDataSource reads an Ancla service.
type ServiceDataSource struct {
	client *client.Client
}

// ServiceDataSourceModel maps the data source schema data.
type ServiceDataSourceModel struct {
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

func NewServiceDataSource() datasource.DataSource {
	return &ServiceDataSource{}
}

func (d *ServiceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

func (d *ServiceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an Ancla service by workspace, project, environment, and service slug.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the service.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The display name of the service.",
				Computed:    true,
			},
			"slug": schema.StringAttribute{
				Description: "The URL-friendly slug of the service.",
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
			"env_slug": schema.StringAttribute{
				Description: "The slug of the environment.",
				Required:    true,
			},
			"platform": schema.StringAttribute{
				Description: "The platform type of the service.",
				Computed:    true,
			},
			"github_repository": schema.StringAttribute{
				Description: "The GitHub repository linked to this service.",
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

func (d *ServiceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServiceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ServiceDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	svc, err := d.client.GetService(
		config.WorkspaceSlug.ValueString(),
		config.ProjectSlug.ValueString(),
		config.EnvSlug.ValueString(),
		config.Slug.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error reading service", err.Error())
		return
	}

	config.ID = types.StringValue(svc.ID)
	config.Name = types.StringValue(svc.Name)
	config.Slug = types.StringValue(svc.Slug)
	config.Platform = types.StringValue(svc.Platform)
	config.GithubRepository = types.StringValue(svc.GithubRepository)
	config.AutoDeployBranch = types.StringValue(svc.AutoDeployBranch)

	if len(svc.ProcessCounts) > 0 {
		elems := make(map[string]types.Int64, len(svc.ProcessCounts))
		for k, v := range svc.ProcessCounts {
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
