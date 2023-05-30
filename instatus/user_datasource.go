package instatus

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	is "github.com/paydaycay/instatus-client-go"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &userDataSource{}
	_ datasource.DataSourceWithConfigure = &userDataSource{}
)

// NewUserDataSource is a helper function to simplify the provider implementation.
func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

// userDataSource is the data source implementation.
type userDataSource struct {
	client *is.Client
}

// userDataSourceModel maps the data source schema data.
type userDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	Email  types.String `tfsdk:"email"`
	Name   types.String `tfsdk:"name"`
	Slug   types.String `tfsdk:"slug"`
	Avatar types.String `tfsdk:"avatar"`
}

// Metadata returns the data source type name.
func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the schema for the data source.
func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves user profile.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the user.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the user.",
				Computed:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email of the user.",
				Computed:    true,
			},
			"slug": schema.StringAttribute{
				Description: "Slug of the user.",
				Computed:    true,
			},
			"avatar": schema.StringAttribute{
				Description: "Avatar url of the user.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*is.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state userDataSourceModel

	user, err := d.client.GetUser()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Instatus User",
			err.Error(),
		)
		return
	}

	// Map response body to model
	state.ID = types.StringValue(user.ID)
	state.Name = types.StringValue(user.Name)
	state.Email = types.StringValue(user.Email)
	state.Slug = types.StringValue(user.Slug)
	state.Avatar = types.StringValue(user.Avatar)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
