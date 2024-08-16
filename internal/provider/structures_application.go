package provider

import (
	"context"
	"fmt"

	"github.com/ans-group/sdk-go/pkg/service/account"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func expandApplicationScope(ctx context.Context, rawAppScope []ApplicationScopeModel) []account.ApplicationServiceScope {
	appScope := make([]account.ApplicationServiceScope, len(rawAppScope))

	for i := range rawAppScope {
		appScope[i] = account.ApplicationServiceScope{
			Service: rawAppScope[i].Service.ValueString(),
			Roles:   expandArray(ctx, rawAppScope[i].Roles),
		}
	}

	tflog.Info(ctx, fmt.Sprintf("Application scopes: %+v", appScope))

	return appScope
}

func expandArray(ctx context.Context, rawArray []types.String) []string {
	expandedArray := make([]string, len(rawArray))

	for i, v := range rawArray {
		expandedArray[i] = v.String()
	}

	tflog.Info(ctx, fmt.Sprintf("Expanded array data: %+v", expandedArray))

	return expandedArray
}

func readApplicationScope(ctx context.Context, rawAppScope []account.ApplicationServiceScope) []ApplicationScopeModel {
	appScope := make([]ApplicationScopeModel, len(rawAppScope))

	for i := range rawAppScope {
		appScope[i] = ApplicationScopeModel{
			Service: types.StringValue(rawAppScope[i].Service),
			Roles:   readAppScopeArray(ctx, rawAppScope[i].Roles),
		}
	}

	tflog.Info(ctx, fmt.Sprintf("Application scopes: %+v", appScope))

	return appScope
}

func readAppScopeArray(ctx context.Context, rawArray []string) []types.String {
	expandedArray := make([]types.String, len(rawArray))

	for i, v := range rawArray {
		expandedArray[i] = types.StringValue(v)
	}

	tflog.Info(ctx, fmt.Sprintf("Expanded array data: %+v", expandedArray))

	return expandedArray
}
