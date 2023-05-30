package instatus

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	is "github.com/paydaycay/instatus-client-go"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &instatusProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New() provider.Provider {
	return &instatusProvider{}
}

// instatusProvider is the provider implementation.
type instatusProvider struct{}

type instatusProviderModel struct {
	ApiKey types.String `tfsdk:"api_key"`
}

// Metadata returns the provider type name.
func (p *instatusProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "instatus"
}

// Schema defines the provider-level schema for configuration data.
func (p *instatusProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Instatus.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "API Key for Instatus API. May also be provided via INSTATUS_APIKEY environment variable.",
				Required:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure prepares a Instatus API client for data sources and resources.
func (p *instatusProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config instatusProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.ApiKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Instatus API Key",
			"The provider cannot create the Instatus API client as there is an unknown configuration value for the Instatus API Key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the INSTATUS_APIKEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	apiKey := os.Getenv("INSTATUS_APIKEY")

	if !config.ApiKey.IsNull() {
		apiKey = config.ApiKey.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("apiKey"),
			"Missing Instatus API Key",
			"The provider cannot create the Instatus API client as there is a missing or empty value for the Instatus API Key. "+
				"Set the host value in the configuration or use the HASHICUPS_APIKEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new Instatus client using the configuration values
	client := is.NewClient(apiKey)

	// Make the Instatus client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *instatusProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewUserDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *instatusProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewComponentResource,
		NewTemplateResource,
	}
}
