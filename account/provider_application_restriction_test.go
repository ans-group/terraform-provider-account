package provider

import (
	"fmt"
	"testing"

	accountservice "github.com/ans-group/sdk-go/pkg/service/account"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccApplicationRestriction_basic(t *testing.T) {
	resourceName := "account_application_restriction.test-application-restriction"

	service := AccTestingClient{}
	service.Configure()
	restrictionType := "allowlist"
	ipRange := "1.1.1.1/24"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             service.testAccCheckApplicationRestrictionDestroy,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testAccResourceApplicationRestrictionConfig_basic(restrictionType, ipRange),
				Check: resource.ComposeTestCheckFunc(
					service.testAccCheckApplicationRestrictionExists(t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "type", restrictionType),
					resource.TestCheckResourceAttr(resourceName, "ranges.0", ipRange),
				),
			},
		},
	})

	restrictionType = "denylist"
	ipRange = "1.1.1.1"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             service.testAccCheckApplicationRestrictionDestroy,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testAccResourceApplicationRestrictionConfig_basic(restrictionType, ipRange),
				Check: resource.ComposeTestCheckFunc(
					service.testAccCheckApplicationRestrictionExists(t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "type", restrictionType),
					resource.TestCheckResourceAttr(resourceName, "ranges.0", ipRange),
				),
			},
		},
	})

}

func (r *AccTestingClient) testAccCheckApplicationRestrictionExists(t *testing.T, n string) resource.TestCheckFunc {
	service := r.client
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			t.Fatalf("Application not found: %s", n)
		}

		if rs.Primary.ID == "" {
			t.Fatal("")
		}

		_, err := service.GetApplicationRestrictions(rs.Primary.Attributes["application_id"])
		if err != nil {
			if _, ok := err.(*accountservice.ApplicationNotFoundError); ok {
				t.Fatal("Application Not found")
			}
			t.Fatal(err)
		}

		return nil
	}
}

func (r *AccTestingClient) testAccCheckApplicationRestrictionDestroy(s *terraform.State) error {
	service := r.client
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "account_application" {
			continue
		}
		if rs.Primary.Attributes["application_id"] != "" {
			_, err := service.GetApplicationRestrictions(rs.Primary.Attributes["application_id"])
			if err == nil {
				return fmt.Errorf("Application with ID [%s] still exists", rs.Primary.ID)
			}

			if _, ok := err.(*accountservice.ApplicationNotFoundError); ok {
				return nil
			}

			return fmt.Errorf("Application with ID [%s] still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccResourceApplicationRestrictionConfig_basic(restrictionType string, ipRange string) string {
	return fmt.Sprintf(`
		resource "account_application" "test-application"{
			name = "tftest-application-restriction"
			description = "test"
		}

		resource "account_application_restriction" "test-application-restriction"{
			application_id = account_application.test-application.id
			type = "%[1]s"
			ranges = ["%[2]s"]
		}
		`, restrictionType, ipRange,
	)
}
