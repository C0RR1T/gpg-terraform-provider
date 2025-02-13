package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testCase = `
	resource "gpg_key_pair" "test" {
		passphrase = "Hello World"
		identity {
			email = "hello@example.com"
			name = "Hello World"
		}
	}
`

func TestAccResource(t *testing.T) {
	dsn := "gpg_key_pair.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testCase,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dsn, "id"),
				),
			},
		},
	})
}
