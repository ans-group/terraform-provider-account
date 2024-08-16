// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"

	accountservice "github.com/ans-group/sdk-go/pkg/service/account"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AccountsApplication{}
var _ resource.ResourceWithImportState = &AccountsApplication{}

func NewAccountsApplication() resource.Resource {
	return &AccountsApplication{}
}

// AccountsApplication defines the resource implementation.
type AccountsApplication struct {
	client *http.Client
}

// AccountsApplicationModel describes the resource data model.
type AccountsApplicationModel struct {
	ID            types.String                  `tfsdk:"id"`
	Name          types.String                  `tfsdk:"name"`
	Description   types.String                  `tfsdk:"description"`
	Scope         []ApplicationScopeModel       `tfsdk:"scope"`
	IPRestriction ApplicationIPRestrictionModel `tfsdk:"ip_restriction"`
}

type ApplicationScopeModel struct {
	Service types.String   `tfsdk:"service"`
	Roles   []types.String `tfsdk:"roles"`
}

type ApplicationIPRestrictionModel struct {
	Type   types.String   `tfsdk:"type"`
	Ranges []types.String `tfsdk:"ranges"`
}

func (r *AccountsApplication) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

// Schema defines the schema for the resource.
func (r *AccountsApplication) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"scope": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"service": schema.StringAttribute{
							Required: true,
						},
						"roles": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
						},
					},
				},
			},
			"ip_restriction": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required: true,
					},
					"ranges": schema.ListAttribute{
						ElementType: types.StringType,
						Required:    true,
					},
				},
			},
		},
	}
}

func (r *AccountsApplication) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var d AccountsApplicationModel
	var meta interface{}
	service := meta.(accountservice.AccountService)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &d)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := accountservice.CreateApplicationRequest{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
	}

	tflog.Debug(ctx, fmt.Sprintf("Created AccountsApplicationAPIModel: %+v", createReq))

	tflog.Info(ctx, "Creating API Application")
	createData, err := service.CreateApplication(createReq)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error creating network rule: %s", err))
		return
	}

	if d.Scope != nil {
		tflog.Info(ctx, "Setting API Application Services")
		setServiceReq := accountservice.SetServiceRequest{
			Scopes: expandApplicationScope(ctx, d.Scope),
		}

		err = service.SetApplicationServices(createData.ID, setServiceReq)

		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Error Setting application services: %s", err))
			return
		}
	}

	if d.IPRestriction.Ranges != nil {
		tflog.Info(ctx, "Setting API Application Restriction")
		setRestrictionReq := accountservice.SetRestrictionRequest{
			IPRestrictionType: d.IPRestriction.Type.ValueString(),
			IPRanges:          expandArray(ctx, d.IPRestriction.Ranges),
		}

		err = service.SetApplicationRestrictions(createData.ID, setRestrictionReq)

		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Error Setting application restrictions: %s", err))
			return
		}
	}

	d.ID = types.StringValue(createData.ID)

	tflog.Trace(ctx, "created a resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &d)...)
}

func (r *AccountsApplication) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var d AccountsApplicationModel
	var meta interface{}
	service := meta.(accountservice.AccountService)
	resp.Diagnostics.Append(req.State.Get(ctx, &d)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Retrieving API Application", map[string]interface{}{
		"id": d.ID.ValueString(),
	})

	application, err := service.GetApplication(d.ID.ValueString())

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error Retrieving Application: %s", err))
		return
	}

	d.Name = types.StringValue(application.Name)
	d.Description = types.StringValue(application.Description)

	tflog.Info(ctx, "Retrieving API Application Services", map[string]interface{}{
		"id": d.ID.ValueString(),
	})

	services, err := service.GetApplicationServices(d.ID.ValueString())

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error Retrieving Application Scope: %s", err))
		return
	}

	d.Scope = readApplicationScope(ctx, services.Scopes)

	tflog.Info(ctx, "Retrieving API Application Restrictions", map[string]interface{}{
		"id": d.ID.ValueString(),
	})

	restrictions, err := service.GetApplicationRestrictions(d.ID.ValueString())

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error Retrieving Application Restrictions: %s", err))
		return
	}

	d.IPRestriction.Type = types.StringValue(restrictions.IPRestrictionType)
	d.IPRestriction.Ranges = readAppScopeArray(ctx, restrictions.IPRanges)

	resp.Diagnostics.Append(resp.State.Set(ctx, &d)...)
}

func (r *AccountsApplication) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, d AccountsApplicationModel
	var meta interface{}
	service := meta.(accountservice.AccountService)

	if !plan.Name.Equal(d.Name) || !plan.Description.Equal(d.Description) {
		tflog.Info(ctx, "Updating Application Key Details", map[string]interface{}{
			"id": d.ID.ValueString(),
		})

		updateReq := accountservice.UpdateApplicationRequest{
			Name:        d.Name.ValueString(),
			Description: d.Description.ValueString(),
		}

		err := service.UpdateApplication(d.ID.ValueString(), updateReq)

		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Error Updating Application Details: %s", err))
			return
		}
	}

	if plan.Scope != nil {
		tflog.Info(ctx, "Setting API Application Services")
		setServiceReq := accountservice.SetServiceRequest{
			Scopes: expandApplicationScope(ctx, d.Scope),
		}

		err := service.SetApplicationServices(d.ID.ValueString(), setServiceReq)

		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Error Setting application services: %s", err))
			return
		}
	}

	if plan.IPRestriction.Ranges != nil {
		tflog.Info(ctx, "Setting API Application Restriction")
		setRestrictionReq := accountservice.SetRestrictionRequest{
			IPRestrictionType: d.IPRestriction.Type.ValueString(),
			IPRanges:          expandArray(ctx, d.IPRestriction.Ranges),
		}

		err := service.SetApplicationRestrictions(d.ID.ValueString(), setRestrictionReq)

		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Error Setting application restrictions: %s", err))
			return
		}
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &d)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &d)...)
}

func (r *AccountsApplication) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var d AccountsApplicationModel
	var meta interface{}
	service := meta.(accountservice.AccountService)

	resp.Diagnostics.Append(req.State.Get(ctx, &d)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Removing API Application")

	id := d.ID.ValueString()

	err := service.DeleteApplication(id)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error deleting application: %s", err))
		return
	}
}

func (r *AccountsApplication) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
