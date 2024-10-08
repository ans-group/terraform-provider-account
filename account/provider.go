// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"terraform-provider-account/pkg/logger"

	"github.com/ans-group/sdk-go/pkg/client"
	"github.com/ans-group/sdk-go/pkg/config"
	"github.com/ans-group/sdk-go/pkg/connection"
	"github.com/ans-group/sdk-go/pkg/logging"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const userAgent = "terraform-provider-account"

var (
	_ provider.Provider = &accountProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &accountProvider{
			version: version,
		}
	}
}

type accountProvider struct {
	version string
}

type accountProviderModel struct {
	Context types.String `tfsdk:"context"`
	APIKey  types.String `tfsdk:"api_key"`
}

func (p *accountProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "account"
	resp.Version = p.version
}

func (p *accountProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"context": schema.StringAttribute{
				Optional:    true,
				Description: "Config context to use",
			},
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "API token to authenticate with UKFast APIs. See https://developers.ukfast.io for more details",
			},
		},
		Description: "Official ANS Account Terraform provider, allowing for manipulation of Glass Account environments",
	}
}

func (p *accountProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var configuration accountProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &configuration)...)

	err := config.Init("")
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("context"), "Failed to initialise config: ", err.Error(),
		)
	}

	if config.GetBool("api_debug") {
		logging.SetLogger(&logger.ProviderLogger{})
	}

	tflog.Debug(ctx, fmt.Sprintf("Created config model: %+v", configuration))

	context := configuration.Context.ValueString()
	if len(context) > 0 {
		err := config.SwitchCurrentContext(context)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("context"), "Error: ", err.Error(),
			)
		}
	}

	apiKey := configuration.APIKey.ValueString()
	if len(apiKey) > 0 {
		config.Set(config.GetCurrentContextName(), "api_key", apiKey)
	}

	conn, err := getConnection()
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("connection"), "Error: ", err.Error(),
		)
	}

	diags := req.Config.Get(ctx, &configuration)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := client.NewClient(conn).AccountService()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Account API Client",
			"An unexpected error occurred when creating the Account API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Account Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func getConnection() (connection.Connection, error) {
	connFactory := connection.NewDefaultConnectionFactory(
		connection.WithDefaultConnectionUserAgent(userAgent),
	)

	return connFactory.NewConnection()
}

// DataSources defines the data sources implemented in the provider.
func (p *accountProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

// Resources defines the resources implemented in the provider.
func (p *accountProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAccountApplication,
		NewApplicationIPRestriction,
		NewApplicationServiceMapping,
	}
}
