terraform {
  required_providers {
    zerossl = {
      source = "toowoxx/zerossl"
    }
  }
}

resource "zerossl_eab_credentials" "eab_credentials" {
  // Replace with your own API key (https://app.zerossl.com/developer)
  api_key = "0123456789abcdef0123456789abcdef"
}

output "kid" {
  value     = zerossl_eab_credentials.eab_credentials.kid
  sensitive = true
}

output "hmac_key" {
  value     = zerossl_eab_credentials.eab_credentials.hmac_key
  sensitive = true
}
