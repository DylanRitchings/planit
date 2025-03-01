variable "TENANCY_OCID" {type = string}
variable "USER_OCID" {type = string}
variable "FINGERPRINT" {type = string}
variable "COMPARTMENT_OCID" {type = string}
variable "SUBNET_ID" { type = string}
variable "VCN_ID" { type = string}


variable "region" {
  type = string
  default = "uk-london-1"
}

variable "availability_domain" {
  type = string
  default = "xFsp:UK-LONDON-1-AD-3"
}

variable "ssh_public_key" {
  default = "~/.ssh/id_rsa.pub"
  type = string
  }

variable "ssh_private_key" {
  default = "~/.ssh/id_rsa"
  type = string
  }

variable "server_dir" {
  default = "../server/"
  type = string
}

variable "CLOUDFLARE_ZONE_ID" {type = string}
variable "CLOUDFLARE_API_KEY" {type = string}




