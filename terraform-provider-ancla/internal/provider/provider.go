// Package provider implements the Ancla Terraform provider.
package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sidequest-labs/terraform-provider-ancla/internal/client"
	"github.com/sidequest-labs/terraform-provider-ancla/internal/resources"
	datasources "github.com/sidequest-labs/terraform-provider-ancla/internal/resources/datasources"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &AnclaProvider{}
)

// AnclaProvider is the provider implementation.
type AnclaProvider struct {
	version string
}

// AnclaProviderModel maps provider schema data to a Go type.
type AnclaProviderModel struct {
	Server types.String `tfsdk:"server"`
	APIKey types.String `tfsdk:"api_key"`
}

// New returns a function that creates new provider instances.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AnclaProvider{
			version: version,
		}
	}
}

func (p *AnclaProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ancla"
	resp.Version = p.version
}

func (p *AnclaProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Ancla provider is used to manage resources on the Ancla PaaS platform.",
		Attributes: map[string]schema.Attribute{
			"server": schema.StringAttribute{
				Description: "The Ancla server URL. Defaults to https://ancla.dev. Can also be set with the ANCLA_SERVER environment variable.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "The API key for authentication. Can also be set with the ANCLA_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *AnclaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config AnclaProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve server URL: config > env > default.
	server := "https://ancla.dev"
	if !config.Server.IsNull() && !config.Server.IsUnknown() {
		server = config.Server.ValueString()
	} else if v := os.Getenv("ANCLA_SERVER"); v != "" {
		server = v
	}

	// Resolve API key: config > env.
	apiKey := ""
	if !config.APIKey.IsNull() && !config.APIKey.IsUnknown() {
		apiKey = config.APIKey.ValueString()
	} else if v := os.Getenv("ANCLA_API_KEY"); v != "" {
		apiKey = v
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The provider requires an api_key to be set in the provider configuration or via the ANCLA_API_KEY environment variable.",
		)
		return
	}

	c := client.New(server, apiKey)
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *AnclaProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewWorkspaceResource,
		resources.NewProjectResource,
		resources.NewEnvironmentResource,
		resources.NewServiceResource,
		resources.NewConfigResource,
	}
}

func (p *AnclaProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewWorkspaceDataSource,
		datasources.NewProjectDataSource,
		datasources.NewEnvironmentDataSource,
		datasources.NewServiceDataSource,
	}
}
