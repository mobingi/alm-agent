#!/bin/bash

yum install -y git gcc tmux amazon-efs-utils docker

archive=go1.11.1.linux-amd64.tar.gz
if [ ! -f $archive ]; then
  wget -q https://storage.googleapis.com/golang/$archive
fi

if [ ! -d /home/ec2-user/go ]; then
  tar -C /home/ec2-user -xzf $archive
  chown -R ec2-user.ec2-user /home/ec2-user/go
fi

if ! grep GOPATH /home/ec2-user/.bashrc; then
  echo 'export GOROOT=/home/ec2-user/go' >> /home/ec2-user/.bashrc
  echo 'export PATH=$PATH:$GOROOT/bin:/home/ec2-user/bin' >> /home/ec2-user/.bashrc
  echo 'export GOPATH=/home/ec2-user' >> /home/ec2-user/.bashrc
  echo 'sudo chown -R ec2-user.ec2-user /home/ec2-user/src' >> /home/ec2-user/.bashrc
fi

# for efs support development

mkdir -p /mnt/shared
