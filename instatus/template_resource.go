package instatus

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	is "github.com/paydaycay/instatus-client-go"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &templateResource{}
	_ resource.ResourceWithConfigure   = &templateResource{}
	_ resource.ResourceWithImportState = &templateResource{}
)

// Configure adds the provider configured client to the resource.
func (r *templateResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*is.Client)
}

// NewTemplateResource is a helper function to simplify the provider implementation.
func NewTemplateResource() resource.Resource {
	return &templateResource{}
}

// templateResource is the resource implementation.
type templateResource struct {
	client *is.Client
}

// templateResourceModel maps the resource schema data.
type templateResourceModel struct {
	ID          types.String             `tfsdk:"id"`
	PageID      types.String             `tfsdk:"page_id"`
	Subdomain   types.String             `tfsdk:"subdomain"`
	Name        types.String             `tfsdk:"name"`
	Type        types.String             `tfsdk:"type"`
	Message     types.String             `tfsdk:"message"`
	Status      types.String             `tfsdk:"status"`
	Components  []templateComponentModel `tfsdk:"components"`
	Notify      types.Bool               `tfsdk:"notify"`
	LastUpdated types.String             `tfsdk:"last_updated"`
}

type templateComponentModel struct {
	ID     types.String `tfsdk:"id"`
	Status types.String `tfsdk:"status"`
}

// Metadata returns the resource type name.
func (r *templateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template"
}

// Schema defines the schema for the resource.
func (r *templateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	componentStatuses := []string{"OPERATIONAL", "UNDERMAINTENANCE", "DEGRADEDPERFORMANCE", "PARTIALOUTAGE", "MAJOROUTAGE"}
	templateStatuses := []string{"INVESTIGATING", "IDENTIFIED", "MONITORING", "RESOLVED", "NOTSTARTEDYET", "INPROGRESS", "COMPLETED"}
	templateTypes := []string{"MAINTENANCE", "INCIDENT"}

	resp.Schema = schema.Schema{
		Description: "Manages a template.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "String Identifier of the template.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"page_id": schema.StringAttribute{
				Description: "String Identifier of the page of the template.",
				Required:    true,
			},
			"subdomain": schema.StringAttribute{
				Description: "Subdomain of the page of the template.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: fmt.Sprintf("Type of the template. One of: (%s).", strings.Join(templateTypes, ", ")),
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf(templateTypes...)},
			},
			"status": schema.StringAttribute{
				Description: fmt.Sprintf("Status of the template. One of: (%s).", strings.Join(templateStatuses, ", ")),
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf(templateStatuses...)},
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the template.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the template.",
				Required:    true,
			},
			"message": schema.StringAttribute{
				Description: "Message of the template.",
				Required:    true,
			},
			"notify": schema.BoolAttribute{
				Description: "Whether notify is enabled for the template.",
				Optional:    true,
			},
			"components": schema.ListNestedAttribute{
				Description: "List of components in the template with their status.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "String Identifier of the component.",
							Required:    true,
						},
						"status": schema.StringAttribute{
							Description: fmt.Sprintf("Status of the component. One of: (%s).", strings.Join(componentStatuses, ", ")),
							Required:    true,
							Validators:  []validator.String{stringvalidator.OneOf(componentStatuses...)},
						},
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *templateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan templateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var components []is.TemplateComponent
	for _, component := range plan.Components {
		components = append(components, is.TemplateComponent{
			ID:     component.ID.ValueStringPointer(),
			Status: component.Status.ValueStringPointer(),
		})
	}

	var item is.Template = is.Template{
		Name:       plan.Name.ValueStringPointer(),
		Type:       plan.Type.ValueStringPointer(),
		Subdomain:  plan.Subdomain.ValueStringPointer(),
		Message:    plan.Message.ValueStringPointer(),
		Status:     plan.Status.ValueStringPointer(),
		Notify:     plan.Notify.ValueBoolPointer(),
		Components: components,
	}

	// Create new template
	template, err := r.client.CreateTemplate(plan.PageID.ValueString(), &item)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating template",
			"Could not create template, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringPointerValue(template.ID)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *templateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state templateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed template value from Instatus
	template, err := r.client.GetTemplate(state.PageID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Instatus Template",
			"Could not read Instatus template ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	// Overwrite items with refreshed state
	state.Name = types.StringPointerValue(template.Name)
	state.Type = types.StringPointerValue(template.Type)
	state.Message = types.StringPointerValue(template.Message)
	state.Status = types.StringPointerValue(template.Status)
	state.Notify = types.BoolPointerValue(template.Notify)
	state.Components = []templateComponentModel{}
	for _, component := range template.Components {
		state.Components = append(state.Components, templateComponentModel{
			ID:     types.StringPointerValue(component.ComponentID),
			Status: types.StringPointerValue(component.Status),
		})
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *templateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan templateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var components []is.TemplateComponent
	for _, component := range plan.Components {
		components = append(components, is.TemplateComponent{
			ID:     component.ID.ValueStringPointer(),
			Status: component.Status.ValueStringPointer(),
		})
	}

	var item is.Template = is.Template{
		Name:       plan.Name.ValueStringPointer(),
		Subdomain:  plan.Subdomain.ValueStringPointer(),
		Type:       plan.Type.ValueStringPointer(),
		Message:    plan.Message.ValueStringPointer(),
		Status:     plan.Status.ValueStringPointer(),
		Notify:     plan.Notify.ValueBoolPointer(),
		Components: components,
	}

	// Update existing template
	_, err := r.client.UpdateTemplate(plan.PageID.ValueString(), plan.ID.ValueString(), &item)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Instatus Template",
			"Could not update template, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *templateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state templateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing template
	err := r.client.DeleteTemplate(state.PageID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Instatus Template",
			"Could not delete template, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *templateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
