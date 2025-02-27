
apply:
  source .env && cd terraform && terraform apply
  

init:
  source .env && cd terraform && terraform init
