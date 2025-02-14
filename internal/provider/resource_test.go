package provider

import (
	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

func checkIsPrivateKey() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["gpg_key_pair.test"]
		if !ok {
			return fmt.Errorf("Could not find resource gpg_key_pair.test")
		}

		key, err := crypto.NewKeyFromArmored(rs.Primary.Attributes["private_key"])
		if err != nil {
			return err
		}

		if !key.IsPrivate() {
			return fmt.Errorf("Private key is not private key")
		}
	}
}
