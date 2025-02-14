package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/ProtonMail/gopenpgp/v3/crypto"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var testCase = fmt.Sprintf(`
	resource "gpg_key_pair" "test" {
		passphrase = "Hello World"
		identity {
			email = "hello@example.com"
			name = "Hello World"
		}
		expires_at = "%s"
	}
`, getNextYear().Format(time.RFC3339))

func getNextYear() time.Time {
	return time.Now().AddDate(1, 0, 0)
}

func TestAccResource(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testCase,
				Check: resource.ComposeTestCheckFunc(
					checkIsPrivateKey(),
				),
			},
			{
				Config: testCase,
				Check: resource.ComposeTestCheckFunc(
					checkIsPublicKey(),
				),
			},
			{
				Config: testCase,
				Check: resource.ComposeTestCheckFunc(
					checkIdentity(),
				),
			},
			{
				Config: testCase,
				Check: resource.ComposeTestCheckFunc(
					checkExpiration(),
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

		key, err := crypto.NewPrivateKeyFromArmored(rs.Primary.Attributes["private_key"], []byte(rs.Primary.Attributes["passphrase"]))
		if err != nil {
			return err
		}

		if !key.IsPrivate() {
			return fmt.Errorf("Private key is not private")
		}

		locked, err := key.IsLocked()

		if err != nil {
			return err
		}

		if locked {
			return fmt.Errorf("Key is locked")
		}

		return nil
	}
}

func checkIsPublicKey() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["gpg_key_pair.test"]
		if !ok {
			return fmt.Errorf("Could not find resource gpg_key_pair.test")
		}

		key, err := crypto.NewKeyFromArmored(rs.Primary.Attributes["public_key"])
		if err != nil {
			return err
		}

		if key.IsPrivate() {
			return fmt.Errorf("Public key is private")
		}

		return nil
	}
}

func checkIdentity() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["gpg_key_pair.test"]
		if !ok {
			return fmt.Errorf("Could not find resource gpg_key_pair.test")
		}

		key, err := crypto.NewKeyFromArmored(rs.Primary.Attributes["public_key"])
		if err != nil {
			return err
		}

		_, ok = key.GetEntity().Identities["Hello World <hello@example.com>"]

		if !ok {
			return fmt.Errorf("Could not find identity \"Hello World <hello@example.com>\"")
		}

		return nil
	}
}

func checkExpiration() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["gpg_key_pair.test"]
		if !ok {
			return fmt.Errorf("Could not find resource gpg_key_pair.test")
		}

		key, err := crypto.NewKeyFromArmored(rs.Primary.Attributes["public_key"])
		if err != nil {
			return err
		}

		if key.IsExpired(time.Now().Unix()) {
			return fmt.Errorf("Key is expired when it shouldn't be")
		}

		if !key.IsExpired(time.Now().AddDate(2, 0, 0).Unix()) {
			return fmt.Errorf("Key is not expired when it should be")
		}

		return nil
	}
}
