#! /bin/bash

exec > /var/log/startup-script-custom.log 2>&1


mkdir mandelbrot
cd mandelbrot

apt-get update -y   

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

terraform init -input=false
terraform plan -out=tfplan -input=false
terraform apply tfplan
ansible-playbook -i inventory.ini playbook.yml