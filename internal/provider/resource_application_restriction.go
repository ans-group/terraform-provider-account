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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ApplicationIPRestriction{}
var _ resource.ResourceWithImportState = &ApplicationIPRestriction{}

func NewApplicationIPRestriction() resource.Resource {
	return &ApplicationIPRestriction{}
}

// ApplicationApplication defines the resource implementation.
type ApplicationIPRestriction struct {
	client accountservice.AccountService
}

type ApplicationIPRestrictionModel struct {
	ApplicationID types.String   `tfsdk:"application_id"`
	Type          types.String   `tfsdk:"type"`
	Ranges        []types.String `tfsdk:"ranges"`
}

func (r *ApplicationIPRestriction) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_restriction"
}

// Schema defines the schema for the resource.
func (r *ApplicationIPRestriction) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_id": schema.StringAttribute{
				Required: true,
			},
			"type": schema.StringAttribute{
				Required: true,
			},
			"ranges": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
		},
	}
}

func (r *ApplicationIPRestriction) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ApplicationIPRestriction) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var d ApplicationIPRestrictionModel
	service := r.client

	resp.Diagnostics.Append(req.Plan.Get(ctx, &d)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Setting API Application Restriction")
	setRestrictionReq := accountservice.SetRestrictionRequest{
		IPRestrictionType: d.Type.ValueString(),
		IPRanges:          expandArray(ctx, d.Ranges),
	}

	err := service.SetApplicationRestrictions(d.ApplicationID.ValueString(), setRestrictionReq)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Setting Application Restrictions",
			fmt.Sprint(err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &d)...)
}

func (r *ApplicationIPRestriction) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var d ApplicationIPRestrictionModel
	service := r.client

	resp.Diagnostics.Append(req.State.Get(ctx, &d)...)

	if resp.Diagnostics.HasError() {
		return
	}

	restrictions, err := service.GetApplicationRestrictions(d.ApplicationID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Retrieving Application Restrictions",
			fmt.Sprint(err),
		)
		return
	}

	d.Type = types.StringValue(restrictions.IPRestrictionType)
	d.Ranges = readApiArray(ctx, restrictions.IPRanges)
	resp.Diagnostics.Append(resp.State.Set(ctx, &d)...)
}

func (r *ApplicationIPRestriction) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, d ApplicationIPRestrictionModel
	service := r.client

	resp.Diagnostics.Append(req.State.Get(ctx, &d)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Setting API Application Restriction")

	setRestrictionReq := accountservice.SetRestrictionRequest{
		IPRestrictionType: plan.Type.ValueString(),
		IPRanges:          expandArray(ctx, plan.Ranges),
	}

	err := service.SetApplicationRestrictions(plan.ApplicationID.ValueString(), setRestrictionReq)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Setting Application Restrictions",
			fmt.Sprint(err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApplicationIPRestriction) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var d ApplicationIPRestrictionModel
	service := r.client

	resp.Diagnostics.Append(req.State.Get(ctx, &d)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Removing IP Restrictions")

	_, err := service.GetApplication(d.ApplicationID.ValueString())

	if err != nil {
		return
	}

	tflog.Info(ctx, "Application found, removing services")

	err = service.DeleteApplicationRestrictions(d.ApplicationID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Removing Application Restrictions",
			fmt.Sprint(err),
		)
		return
	}
}

func (r *ApplicationIPRestriction) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
