# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.define "default" do |c|
    c.vm.box = 'ubuntu/trusty64'
    c.vm.provision :shell, path: File.expand_path('../vagrant/prov/ub14.sh', __FILE__)
    c.vm.synced_folder '.', '/home/vagrant/src/github.com/mobingilabs/go-modaemon'
  end
end
