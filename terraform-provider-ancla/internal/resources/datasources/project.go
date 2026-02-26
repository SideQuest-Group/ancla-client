package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sidequest-labs/terraform-provider-ancla/internal/client"
)

var _ datasource.DataSource = &ProjectDataSource{}

// ProjectDataSource reads an Ancla project.
type ProjectDataSource struct {
	client *client.Client
}

// ProjectDataSourceModel maps the data source schema data.
type ProjectDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Slug          types.String `tfsdk:"slug"`
	WorkspaceSlug types.String `tfsdk:"workspace_slug"`
	ServiceCount  types.Int64  `tfsdk:"service_count"`
	Created       types.String `tfsdk:"created"`
	Updated       types.String `tfsdk:"updated"`
}

func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

func (d *ProjectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *ProjectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an Ancla project by workspace slug and project slug.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the project.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The display name of the project.",
				Computed:    true,
			},
			"slug": schema.StringAttribute{
				Description: "The URL-friendly slug of the project.",
				Required:    true,
			},
			"workspace_slug": schema.StringAttribute{
				Description: "The slug of the workspace this project belongs to.",
				Required:    true,
			},
			"service_count": schema.Int64Attribute{
				Description: "The number of services in the project.",
				Computed:    true,
			},
			"created": schema.StringAttribute{
				Description: "The creation timestamp of the project.",
				Computed:    true,
			},
			"updated": schema.StringAttribute{
				Description: "The last update timestamp of the project.",
				Computed:    true,
			},
		},
	}
}

func (d *ProjectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ProjectDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := d.client.GetProject(config.WorkspaceSlug.ValueString(), config.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading project", err.Error())
		return
	}

	config.ID = types.StringValue(project.ID)
	config.Name = types.StringValue(project.Name)
	config.Slug = types.StringValue(project.Slug)
	config.WorkspaceSlug = types.StringValue(project.WorkspaceSlug)
	config.ServiceCount = types.Int64Value(int64(project.ServiceCount))
	config.Created = types.StringValue(project.Created)
	config.Updated = types.StringValue(project.Updated)

	diags = resp.State.Set(ctx, config)
	resp.Diagnostics.Append(diags...)
}
