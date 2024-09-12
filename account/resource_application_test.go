package provider

import (
	"fmt"
	"testing"

	accountservice "github.com/ans-group/sdk-go/pkg/service/account"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccApplication_basic(t *testing.T) {
	applicationName := acctest.RandomWithPrefix("tftest")
	applicationDescription := acctest.RandomWithPrefix("tftest")
	resourceName := "account_application.test-application"

	service := AccTestingClient{}
	service.Configure()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             service.testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testAccResourceApplicationConfig_basic(applicationName, applicationDescription),
				Check: resource.ComposeTestCheckFunc(
					service.testAccCheckApplicationExists(t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", applicationName),
					resource.TestCheckResourceAttr(resourceName, "description", applicationDescription),
				),
			},
		},
	})
}

func (r *AccTestingClient) testAccCheckApplicationExists(t *testing.T, n string) resource.TestCheckFunc {
	service := r.client
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			t.Fatalf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			t.Fatal("No Application ID is set")
		}

		_, err := service.GetApplication(rs.Primary.ID)
		if err != nil {
			if _, ok := err.(*accountservice.ApplicationNotFoundError); ok {
				t.Fatal("Application Not found")
			}
			t.Fatal(err)
		}

		return nil
	}
}

func (r *AccTestingClient) testAccCheckApplicationDestroy(s *terraform.State) error {
	service := r.client
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "account_application" {
			continue
		}

		_, err := service.GetApplicationServices(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Application with ID [%s] still exists", rs.Primary.ID)
		}

		if _, ok := err.(*accountservice.ApplicationNotFoundError); ok {
			return nil
		}

		return fmt.Errorf("Application with ID [%s] still exists", rs.Primary.ID)
	}

	return nil
}

func testAccResourceApplicationConfig_basic(applicationName string, applicationDesription string) string {
	return fmt.Sprintf(`
		resource "account_application" "test-application"{
			name = "%[1]s"
			description = "%[2]s"
		}
		resource "account_application_services" "test-application-services" {
			application_id = account_application.test-application.id
			service {
					name = "ecloud"
					roles = [
						"read", "write"
					]
			}
		}
		`, applicationName, applicationDesription,
	)
}
