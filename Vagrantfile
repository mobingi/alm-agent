# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = 'ubuntu/trusty64'

  config.vm.provision :shell, inline: <<-EOF
    echo Installing golang ...
    archive=go1.7.3.linux-amd64.tar.gz
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
    apt-get install -y pkgconf git cmake

    mkdir -p ~vagrant/src/github.com/mobingilabs/go-modaemon
    chown -R vagrant:vagrant ~vagrant/src

    echo Installing libgit2 ...
    if [ ! -f v0.24.2.tar.gz ]; then
      wget -q https://github.com/libgit2/libgit2/archive/v0.24.2.tar.gz
      tar zxvf v0.24.2.tar.gz
      cd libgit2-0.24.2
      mkdir build && cd build
      cmake ..
      cmake --build .
      cmake .. -DCMAKE_INSTALL_PREFIX=/usr
      cmake --build . --target install
    fi
  EOF

   config.vm.synced_folder '.', '/home/vagrant/src/github.com/mobingilabs/go-modaemon'
end
