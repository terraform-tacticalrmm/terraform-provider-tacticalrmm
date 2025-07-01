package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &trmmProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &trmmProvider{
			version: version,
		}
	}
}

// trmmProvider is the provider implementation.
type trmmProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// trmmProviderModel describes the provider data model.
type trmmProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	APIKey   types.String `tfsdk:"api_key"`
}

// Metadata returns the provider type name.
func (p *trmmProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "tacticalrmm"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *trmmProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Tactical RMM provider allows you to manage Tactical RMM resources.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "The Tactical RMM API endpoint. Can also be set via TRMM_ENDPOINT environment variable.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "The Tactical RMM API key. Can also be set via TRMM_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure prepares a Tactical RMM API client for data sources and resources.
func (p *trmmProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config trmmProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	endpoint := config.Endpoint.ValueString()
	apiKey := config.APIKey.ValueString()

	// If values aren't known, check environment variables
	if endpoint == "" {
		endpoint = "https://api.tactical-rmm.com" // Default endpoint
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The provider cannot create the Tactical RMM API client as there is a missing or empty value for the API key. "+
				"Set the api_key value in the configuration or use the TRMM_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
		return
	}

	// Create HTTP client
	client := &http.Client{}

	// Create custom client configuration
	clientConfig := &ClientConfig{
		BaseURL:    endpoint,
		APIKey:     apiKey,
		HTTPClient: client,
	}

	// Make the client available to resources and data sources
	resp.DataSourceData = clientConfig
	resp.ResourceData = clientConfig
}

// DataSources defines the data sources implemented in the provider.
func (p *trmmProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Singular data sources (lookup by ID or name)
		NewScriptDataSource,
		NewScriptSnippetDataSource,
		NewKeyStoreDataSource,
		// Plural data sources (list all or filter)
		NewScriptsDataSource,
		NewScriptSnippetsDataSource,
		NewKeyStoresDataSource,
		// Add more data sources here as needed
		// NewAgentsDataSource,
		// NewClientsDataSource,
		// NewSitesDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *trmmProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewScriptResource,
		NewScriptSnippetResource,
		NewKeyStoreResource,
		// NewAgentResource,
		// NewCheckResource,
		// NewTaskResource,
		// NewPolicyResource,
		// NewAlertTemplateResource,
	}
}

// ClientConfig holds the configuration for the TRMM API client
type ClientConfig struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// Do performs an HTTP request with authentication
func (c *ClientConfig) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-API-KEY", c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	return c.HTTPClient.Do(req)
}
