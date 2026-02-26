package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sidequest-labs/terraform-provider-ancla/internal/client"
)

var _ datasource.DataSource = &WorkspaceDataSource{}

// WorkspaceDataSource reads an Ancla workspace.
type WorkspaceDataSource struct {
	client *client.Client
}

// WorkspaceDataSourceModel maps the data source schema data.
type WorkspaceDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Slug         types.String `tfsdk:"slug"`
	MemberCount  types.Int64  `tfsdk:"member_count"`
	ProjectCount types.Int64  `tfsdk:"project_count"`
	ServiceCount types.Int64  `tfsdk:"service_count"`
}

func NewWorkspaceDataSource() datasource.DataSource {
	return &WorkspaceDataSource{}
}

func (d *WorkspaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (d *WorkspaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an Ancla workspace by slug.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the workspace.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The display name of the workspace.",
				Computed:    true,
			},
			"slug": schema.StringAttribute{
				Description: "The URL-friendly slug of the workspace.",
				Required:    true,
			},
			"member_count": schema.Int64Attribute{
				Description: "The number of members in the workspace.",
				Computed:    true,
			},
			"project_count": schema.Int64Attribute{
				Description: "The number of projects in the workspace.",
				Computed:    true,
			},
			"service_count": schema.Int64Attribute{
				Description: "The total number of services across all projects.",
				Computed:    true,
			},
		},
	}
}

func (d *WorkspaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *WorkspaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config WorkspaceDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ws, err := d.client.GetWorkspace(config.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading workspace", err.Error())
		return
	}

	config.ID = types.StringValue(ws.ID)
	config.Name = types.StringValue(ws.Name)
	config.Slug = types.StringValue(ws.Slug)
	config.MemberCount = types.Int64Value(int64(ws.MemberCount))
	config.ProjectCount = types.Int64Value(int64(ws.ProjectCount))
	config.ServiceCount = types.Int64Value(int64(ws.ServiceCount))

	diags = resp.State.Set(ctx, config)
	resp.Diagnostics.Append(diags...)
}
