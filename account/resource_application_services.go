// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	accountservice "github.com/ans-group/sdk-go/pkg/service/account"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ApplicationServiceMapping{}
var _ resource.ResourceWithImportState = &ApplicationServiceMapping{}

func NewApplicationServiceMapping() resource.Resource {
	return &ApplicationServiceMapping{}
}

// ApplicationServiceMapping defines the resource implementation.
type ApplicationServiceMapping struct {
	client accountservice.AccountService
}

type ApplicationServiceMappingModel struct {
	ApplicationID types.String `tfsdk:"application_id"`
	Services      types.List   `tfsdk:"service"`
}

type ApplicationServiceScope struct {
	Name  types.String   `tfsdk:"name"`
	Roles []types.String `tfsdk:"roles"`
}

func (m ApplicationServiceScope) attributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":  types.StringType,
		"roles": types.ListType{types.StringType},
	}
}

func (r *ApplicationServiceMapping) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_services"
}

// Schema defines the schema for the resource.
func (r *ApplicationServiceMapping) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of application to apply services access to",
			},
		},
		Blocks: map[string]schema.Block{
			"service": schema.ListNestedBlock{
				Description: "Defines service access",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of service",
						},
						"roles": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
							Description: "List of service roles",
						},
					},
				},
			},
		},
		Description: "Defines the services which the API key has access to and the access roles it has for each.",
	}
}

func (r *ApplicationServiceMapping) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ApplicationServiceMapping) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var d ApplicationServiceMappingModel
	service := r.client

	resp.Diagnostics.Append(req.Plan.Get(ctx, &d)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Setting API Application Services")

	scopes := make([]ApplicationServiceScope, 0, len(d.Services.Elements()))
	d.Services.ElementsAs(ctx, &scopes, false)

	setServiceReq := accountservice.SetServiceRequest{
		Scopes: expandApplicationScope(ctx, scopes),
	}

	tflog.Info(ctx, fmt.Sprintf("Created Set Service Request: %+v", setServiceReq))

	err := service.SetApplicationServices(d.ApplicationID.ValueString(), setServiceReq)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Setting Application Services",
			fmt.Sprint(err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &d)...)
}

func (r *ApplicationServiceMapping) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var d ApplicationServiceMappingModel
	service := r.client

	resp.Diagnostics.Append(req.State.Get(ctx, &d)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Retrieving API Application Services", map[string]interface{}{
		"id": d.ApplicationID.ValueString(),
	})

	services, err := service.GetApplicationServices(d.ApplicationID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Retrieving Application Services",
			fmt.Sprint(err),
		)
		return
	}

	stateScopes := make([]ApplicationServiceScope, 0, len(d.Services.Elements()))
	d.Services.ElementsAs(ctx, &stateScopes, false)
	expandedStateScopes := expandApplicationScope(ctx, stateScopes)
	readScopes := readApplicationScope(ctx, services.Scopes)

	stateScopesJson, _ := json.Marshal(expandedStateScopes)
	readScopesJson, _ := json.Marshal(services.Scopes)

	if string(stateScopesJson) != string(readScopesJson) {
		d.Services, _ = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: ApplicationServiceScope{}.attributeTypes()}, readScopes)
		resp.Diagnostics.Append(resp.State.Set(ctx, &d)...)
	}
}

func (r *ApplicationServiceMapping) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, d ApplicationServiceMappingModel
	service := r.client

	resp.Diagnostics.Append(req.State.Get(ctx, &d)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	tflog.Info(ctx, "Setting API Application Services")

	scopes := make([]ApplicationServiceScope, 0, len(plan.Services.Elements()))
	plan.Services.ElementsAs(ctx, &scopes, false)

	setServiceReq := accountservice.SetServiceRequest{
		Scopes: expandApplicationScope(ctx, scopes),
	}

	tflog.Info(ctx, fmt.Sprintf("Created Set Service Request: %+v", setServiceReq))

	err := service.SetApplicationServices(d.ApplicationID.ValueString(), setServiceReq)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Setting Application Services",
			fmt.Sprint(err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApplicationServiceMapping) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var d ApplicationServiceMappingModel
	service := r.client

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Removing API Application Services")

	_, err := service.GetApplication(d.ApplicationID.ValueString())

	if err != nil {
		return
	}

	tflog.Info(ctx, "Application found, removing services")

	err = service.DeleteApplicationServices(d.ApplicationID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Removing Application Services",
			fmt.Sprint(err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &d)...)
}

func (r *ApplicationServiceMapping) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
