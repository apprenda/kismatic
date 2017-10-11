#!/bin/bash
sudo apt-get update -y &&
wget --no-check-certificate -O - https://github.com/apprenda/kismatic/releases/download/v1.2.1/kismatic-v1.2.1-linux-amd64.tar.gz | tar -zx && 
sudo apt-get -y install git build-essential &&
sudo apt-get install -qq python2.7 && ln -s /usr/bin/python2.7 /usr/bin/python &&
curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl &&
chmod +x ./kubectl &&
sudo mv ./kubectl /usr/local/bin/kubectl

