#!/bin/bash

echo Installing golang ...
archive=go1.11.1.linux-amd64.tar.gz
if [ ! -f $archive ]; then
  wget -q https://storage.googleapis.com/golang/$archive
fi

if [ ! -d /usr/local/go ]; then
  tar -C /usr/local -xzf $archive
fi

if ! grep GOPATH ~vagrant/.bashrc; then
  echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/bin' >> ~vagrant/.bashrc
  echo 'export GOPATH=$HOME' >> ~vagrant/.bashrc
fi

if [ ! -d ~vagrant/src ]; then
  mkdir ~vagrant/src
  chown vagrant:vagrant ~vagrant/src
fi

echo Installing pkgconf, git and cmake ...
apt-get update
apt-get install -y pkgconf git cmake

mkdir -p ~vagrant/src/github.com/mobingi/alm-agent
chown -R vagrant:vagrant ~vagrant/src

apt-get install -y apt-transport-https ca-certificates
apt-get install -y linux-image-extra-$(uname -r) linux-image-extra-virtual
apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D
echo deb https://apt.dockerproject.org/repo ubuntu-trusty main > /etc/apt/sources.list.d/docker.list
apt-get update
apt-get install -y docker-engine --force-yes
