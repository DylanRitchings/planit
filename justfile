apply:
  source .env && cd terraform && terraform apply -auto-approve
  @sed -i '' '/^export SERVER_IP=/d' .env
  @echo "export SERVER_IP=$(cd terraform && terraform output -raw planit_server_ip)" >> .env

ansible:
  source .env && ansible-playbook -i ${SERVER_IP}, -u ubuntu --private-key ~/.ssh/id_rsa ./ansible/playbook.yml

destroy:
  source .env && cd terraform && terraform destroy

init:
  source .env && cd terraform && terraform init
