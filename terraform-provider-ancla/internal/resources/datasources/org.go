package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sidequest-labs/terraform-provider-ancla/internal/client"
)

var _ datasource.DataSource = &OrgDataSource{}

// OrgDataSource reads an Ancla organization.
type OrgDataSource struct {
	client *client.Client
}

// OrgDataSourceModel maps the data source schema data.
type OrgDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Slug             types.String `tfsdk:"slug"`
	MemberCount      types.Int64  `tfsdk:"member_count"`
	ProjectCount     types.Int64  `tfsdk:"project_count"`
	ApplicationCount types.Int64  `tfsdk:"application_count"`
}

func NewOrgDataSource() datasource.DataSource {
	return &OrgDataSource{}
}

func (d *OrgDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org"
}

func (d *OrgDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an Ancla organization by slug.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the organization.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The display name of the organization.",
				Computed:    true,
			},
			"slug": schema.StringAttribute{
				Description: "The URL-friendly slug of the organization.",
				Required:    true,
			},
			"member_count": schema.Int64Attribute{
				Description: "The number of members in the organization.",
				Computed:    true,
			},
			"project_count": schema.Int64Attribute{
				Description: "The number of projects in the organization.",
				Computed:    true,
			},
			"application_count": schema.Int64Attribute{
				Description: "The total number of applications across all projects.",
				Computed:    true,
			},
		},
	}
}

func (d *OrgDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrgDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config OrgDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := d.client.GetOrg(config.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading organization", err.Error())
		return
	}

	config.ID = types.StringValue(org.ID)
	config.Name = types.StringValue(org.Name)
	config.Slug = types.StringValue(org.Slug)
	config.MemberCount = types.Int64Value(int64(org.MemberCount))
	config.ProjectCount = types.Int64Value(int64(org.ProjectCount))
	config.ApplicationCount = types.Int64Value(int64(org.ApplicationCount))

	diags = resp.State.Set(ctx, config)
	resp.Diagnostics.Append(diags...)
}
