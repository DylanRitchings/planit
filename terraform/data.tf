data "oci_core_images" "latest_image" {
    compartment_id = var.COMPARTMENT_OCID
    operating_system         = "Canonical Ubuntu"
    operating_system_version = "22.04"
    shape                    = "VM.Standard.E2.1.Micro"
  }

