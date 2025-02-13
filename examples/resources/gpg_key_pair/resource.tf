resource "gpg_key_pair" "example" {
  identity {
    email = "hello@example.com"
    name  = "John Doe"
  }

  kind       = "rfc4880"
  passphrase = "Hello world"
  expires_at = time_rotating.example.rotation_rfc3339
}

resource "time_rotating" "example" {
  rotation_years = 1
}
