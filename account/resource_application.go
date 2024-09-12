// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	accountservice "github.com/ans-group/sdk-go/pkg/service/account"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AccountApplication{}
var _ resource.ResourceWithImportState = &AccountApplication{}

func NewAccountApplication() resource.Resource {
	return &AccountApplication{}
}

// AccountApplication defines the resource implementation.
type AccountApplication struct {
	client accountservice.AccountService
}

// AccountApplicationModel describes the resource data model.
type AccountApplicationModel struct {
	ID          types.String `tfsdk:"id"`
	Key         types.String `tfsdk:"key"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (r *AccountApplication) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

// Schema defines the schema for the resource.
func (r *AccountApplication) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *AccountApplication) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(accountservice.AccountService)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected accountservice.AccountService, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *AccountApplication) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var d AccountApplicationModel
	service := r.client

	resp.Diagnostics.Append(req.Plan.Get(ctx, &d)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := accountservice.CreateApplicationRequest{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
	}

	tflog.Debug(ctx, fmt.Sprintf("Created AccountApplicationAPIModel: %+v", createReq))

	tflog.Info(ctx, "Creating API Application")
	createData, err := service.CreateApplication(createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating API Application",
			fmt.Sprint(err),
		)
		return
	}

	d.ID = types.StringValue(createData.ID)
	d.Key = types.StringValue(createData.Key)

	tflog.Trace(ctx, "created a resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &d)...)
}

func (r *AccountApplication) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var d AccountApplicationModel
	service := r.client
	resp.Diagnostics.Append(req.State.Get(ctx, &d)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Retrieving API Application", map[string]interface{}{
		"id": d.ID.ValueString(),
	})

	application, err := service.GetApplication(d.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Retrieving API Application",
			fmt.Sprint(err),
		)
		return
	}

	d.Name = types.StringValue(application.Name)
	d.Description = types.StringValue(application.Description)

	resp.Diagnostics.Append(resp.State.Set(ctx, &d)...)
}

func (r *AccountApplication) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, d AccountApplicationModel
	service := r.client

	resp.Diagnostics.Append(req.State.Get(ctx, &d)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if !plan.Name.Equal(d.Name) || !plan.Description.Equal(d.Description) {
		tflog.Info(ctx, "Updating Application Key Details", map[string]interface{}{
			"id":          plan.ID.ValueString(),
			"name":        plan.Name.ValueString(),
			"description": plan.Description.ValueString(),
		})

		updateReq := accountservice.UpdateApplicationRequest{
			Name:        plan.Name.ValueString(),
			Description: plan.Description.ValueString(),
		}

		err := service.UpdateApplication(d.ID.ValueString(), updateReq)

		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating API Application Details",
				fmt.Sprint(err),
			)
			return
		}
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &d)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &d)...)
}

func (r *AccountApplication) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var d AccountApplicationModel
	service := r.client

	resp.Diagnostics.Append(req.State.Get(ctx, &d)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Removing API Application")

	id := d.ID.ValueString()

	err := service.DeleteApplication(id)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Application",
			fmt.Sprint(err),
		)
		return
	}
}

func (r *AccountApplication) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
