// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/ans-group/sdk-go/pkg/client"
	"github.com/ans-group/sdk-go/pkg/connection"
	accountservice "github.com/ans-group/sdk-go/pkg/service/account"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	providerConfig = (`
	provider "account" {}
	`)
)

type AccTestingClient struct {
	client accountservice.AccountService
}

func (r *AccTestingClient) Configure() {
	apioToken := os.Getenv("APIO_TOKEN_ADMIN")
	conn := connection.NewAPIKeyCredentialsAPIConnection(apioToken)
	c := client.NewClient(conn)

	r.client = c.AccountService()
}

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProvider *schema.Provider
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"account": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	testAccPreCheckRequiredEnvVars(t)
}

func testAccPreCheckRequiredEnvVars(t *testing.T) {
	if os.Getenv("APIO_TOKEN_ADMIN") == "" {
		t.Fatal("APIO_TOKEN_ADMIN must be set for acceptance tests")
	}
}
