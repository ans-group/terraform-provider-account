package provider

import (
	"fmt"
	"testing"

	accountservice "github.com/ans-group/sdk-go/pkg/service/account"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccApplicationService_basic(t *testing.T) {
	resourceName := "account_application_services.test-application-services"

	service := AccTestingClient{}
	service.Configure()
	serviceName := "ecloud"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             service.testAccCheckApplicationServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testAccResourceApplicationServiceConfig_basic(serviceName),
				Check: resource.ComposeTestCheckFunc(
					service.testAccCheckApplicationServiceExists(t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "service.0.name", serviceName),
				),
			},
		},
	})
}

func (r *AccTestingClient) testAccCheckApplicationServiceExists(t *testing.T, n string) resource.TestCheckFunc {
	service := r.client
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			t.Fatalf("Application not found: %s", n)
		}

		if rs.Primary.ID == "" {
			t.Fatal("")
		}

		_, err := service.GetApplicationServices(rs.Primary.Attributes["application_id"])
		if err != nil {
			if _, ok := err.(*accountservice.ApplicationNotFoundError); ok {
				t.Fatal("Application Not found")
			}
			t.Fatal(err)
		}

		return nil
	}
}

func (r *AccTestingClient) testAccCheckApplicationServiceDestroy(s *terraform.State) error {
	service := r.client
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "account_application" {
			continue
		}

		if rs.Primary.Attributes["application_id"] != "" {
			_, err := service.GetApplicationServices(rs.Primary.Attributes["application_id"])

			if err == nil {
				return fmt.Errorf("Application with ID [%s] still exists", rs.Primary.Attributes["application_id"])
			}

			if _, ok := err.(*accountservice.ApplicationNotFoundError); ok {
				return nil
			}

			return fmt.Errorf("Application with ID [%s] still exists", rs.Primary.Attributes["application_id"])
		}

	}

	return nil
}

func testAccResourceApplicationServiceConfig_basic(serviceName string) string {
	return fmt.Sprintf(`
		resource "account_application" "test-application"{
			name = "tftst-app"
			description = "aaa"
		}

		resource "account_application_services" "test-application-services" {
			application_id = account_application.test-application.id
			service {
					name = "%s"
					roles = [
						"read", "write"
					]
			}
		}
		`, serviceName,
	)
}
