package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/ans-group/sdk-go/pkg/service/account"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func expandApplicationScope(ctx context.Context, rawAppScope []ApplicationServiceScope) []account.ApplicationServiceScope {
	appScope := make([]account.ApplicationServiceScope, len(rawAppScope))

	tflog.Info(ctx, fmt.Sprintf("expandApplicationScope Raw Application Service Scope: %+v", rawAppScope))

	for i := range rawAppScope {
		rolesArray := expandArray(ctx, rawAppScope[i].Roles)
		sort.Slice(rolesArray, func(j, k int) bool {
			return rolesArray[j] < rolesArray[k]
		})

		appScope[i] = account.ApplicationServiceScope{
			Service: rawAppScope[i].Name.ValueString(),
			Roles:   rolesArray,
		}
	}

	sort.Slice(appScope[:], func(i, j int) bool {
		return appScope[i].Service < appScope[j].Service
	})

	tflog.Info(ctx, fmt.Sprintf("expandApplicationScope Application Service Scope: %+v", appScope))

	return appScope
}

func expandArray(ctx context.Context, rawArray []types.String) []string {
	expandedArray := make([]string, len(rawArray))

	for i, v := range rawArray {
		expandedArray[i] = v.ValueString()
	}

	tflog.Info(ctx, fmt.Sprintf("Expanded array data: %+v", expandedArray))

	return expandedArray
}

func readApplicationScope(ctx context.Context, rawAppScope []account.ApplicationServiceScope) []ApplicationServiceScope {
	appScope := make([]ApplicationServiceScope, len(rawAppScope))

	tflog.Info(ctx, fmt.Sprintf("Raw Application Service Scope: %+v", rawAppScope))

	sort.Slice(rawAppScope[:], func(i, j int) bool {
		return rawAppScope[i].Service < rawAppScope[j].Service
	})

	for i := range rawAppScope {
		rolesArray := rawAppScope[i].Roles
		sort.Slice(rolesArray, func(j, k int) bool {
			return rolesArray[j] < rolesArray[k]
		})
		appScope[i] = ApplicationServiceScope{
			Name:  types.StringValue(rawAppScope[i].Service),
			Roles: readApiArray(ctx, rolesArray),
		}
	}

	tflog.Info(ctx, fmt.Sprintf("Read Application Service Scope: %+v", appScope))

	return appScope
}

func readApiArray(ctx context.Context, rawArray []string) []types.String {
	expandedArray := make([]types.String, len(rawArray))

	for i, v := range rawArray {
		expandedArray[i] = types.StringValue(v)
	}

	tflog.Info(ctx, fmt.Sprintf("Expanded array data: %+v", expandedArray))

	return expandedArray
}
