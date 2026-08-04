#! /bin/bash

exec > /var/log/startup-script-custom.log 2>&1

# inject env variables
echo "export TF_VAR_client_count=$(curl -H 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/instance/attributes/TF_VAR_client_count)" >> /etc/profile
echo "export TF_VAR_vpc_name=$(curl -H 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/instance/attributes/TF_VAR_vpc_name)" >> /etc/profile
echo "export TF_VAR_subnetwork_name=$(curl -H 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/instance/attributes/TF_VAR_subnetwork_name)" >> /etc/profile
echo "export TF_VAR_project_name=$(curl -H 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/instance/attributes/TF_VAR_project_name)" >> /etc/profile

source /etc/profile

mkdir mandelbrot
cd mandelbrot

apt-get update -y

# terraform dependencies
apt-get install -y gnupg software-properties-common
# Hashicorp key for signing
wget -O- https://apt.releases.hashicorp.com/gpg | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
echo -e "Hashicorop key retrieved.\n"
echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/hashicorp.list
echo -e "Terraform repository added.\n"


# Update the new repository
apt-get update &&
apt-get install -y terraform
echo -e "Terraform installed.\n\n\n"



apt-get install -y git
echo -e "Git installed.\n\n\n"  

mkdir /root/mandelbrot
# Get private key from keystore
gcloud secrets versions access 1 --secret="github_private_key" --project=$TF_VAR_project_name >> /root/mandelbrot/private_key       
chmod 600 /root/mandelbrot/private_key

mkdir /root/.ssh
chmod 700 /root/.ssh

ssh-keyscan github.com >> /root/.ssh/known_hosts 
chmod 600 /root/.ssh/known_hosts

export GIT_SSH_COMMAND="ssh -i /root/mandelbrot/private_key -o BatchMode=yes"

    mkdir repo
    git clone git@github.com:navod-abay/distributed-mandelbrot-go.git repo
    
cd repo/master

pwd

mkdir keys

ssh-keygen -t ed25519 -f keys/client_key -N "" -q

terraform init -input=false
terraform plan -out=tfplan -input=false
terraform apply tfplan 
terraform output -json


apt-get install -y  ansible 
echo -e "Ansible installed.\n\n\n"

ansible-galaxy collection install google.cloud
# Install required Google authentication libraries
pip install google-auth google-cloud-secret-manager

# Sleep to ensure the key is copied from metadata
sleep 5 
ansible-playbook -i inventory.ini playbook.yml