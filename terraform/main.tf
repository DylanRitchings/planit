
terraform {
  required_version = ">= v1.10.1"
 required_providers {

   oci = {
     source = "oracle/oci"
      version = ">= 6.21.0"
   }
 }
}

provider "oci" {
 tenancy_ocid = var.TENANCY_OCID
 user_ocid = var.USER_OCID
 private_key_path = "~/.oci/terraform.pem"
 fingerprint = var.FINGERPRINT
 region = "uk-london-1"
}
