terraform {
  required_version = ">= v1.10.1"
 required_providers {

    oci = {
      source = "oracle/oci"
      version = ">= 6.21.0"
    }

    cloudflare = {
      version = "~> 5"
    }

    archive = {
      source = "hashicorp/archive"
      version = "~> 2.7.0"
    }
  }
}


provider "oci" {
 tenancy_ocid = var.TENANCY_OCID
 user_ocid = var.USER_OCID
 private_key_path = "~/.oci/sessions/DEFAULT/oci_api_key.pem"
 fingerprint = var.FINGERPRINT
 region = var.region
}

provider "cloudflare" {
  api_token = var.CLOUDFLARE_API_KEY  
}


resource "oci_core_instance" "planit_server" {
  compartment_id = var.COMPARTMENT_OCID
  availability_domain = var.availability_domain
  display_name = "planit"
  
  shape = "VM.Standard.E2.1.Micro"

  source_details {
    source_type = "image"
    source_id   = data.oci_core_images.latest_image.images[0].id
  }

  create_vnic_details {
    subnet_id        = var.SUBNET_ID
    assign_public_ip = true
  }

  metadata = {
    ssh_authorized_keys = file(var.ssh_public_key)
  }

}

resource "cloudflare_dns_record" "planit" {
  zone_id = var.CLOUDFLARE_ZONE_ID
  proxied = true
  name    = "planit"  # subdomain 

  type    = "A" 
  content   = oci_core_instance.planit_server.public_ip
  ttl     = 1
}

output "planit_server_ip" {
  value = oci_core_instance.planit_server.public_ip
}
