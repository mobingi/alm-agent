# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.define "default" do |c|
    c.vm.box = 'ubuntu/trusty64'
    c.vm.provision :shell, path: File.expand_path('../vagrant/prov/ub14.sh', __FILE__)
    c.vm.synced_folder '.', '/home/vagrant/src/github.com/mobingi/alm-agent'
  end

  # for plugin development
  config.vm.define "amzn" do |c|
    c.vm.box = 'dummy'
    c.ssh.pty = true

    c.vm.provider :aws do |aws, override|
      aws.aws_profile = ENV['AWS_PROFILE']
      aws.keypair_name = ENV['AWS_KEYPAIR_NAME']
      override.ssh.username = 'ec2-user'
      override.ssh.private_key_path = ENV['AWS_KEYPAIR_PATH']
      # Amazon Linux AMI 2017.03.0.20170401 x86_64 HVM
      aws.region = 'ap-northeast-1'
      aws.region_config 'ap-northeast-1', ami: 'ami-859bbfe2'
      ## << Ref: https://github.com/mitchellh/vagrant-aws/issues/473 aws.region is not loaded from Vagrantfile

      aws.user_data = "#!/bin/bash\nsed -i -e 's/^Defaults.*requiretty/# Defaults requiretty/g' /etc/sudoers"

      aws.instance_type = 't2.medium'
      aws.subnet_id = ENV['AWS_SUBNET_ID']
      aws.associate_public_ip = true
      aws.security_groups = ENV['AWS_SGS'].split(",")
      aws.block_device_mapping = [{ 'DeviceName' => '/dev/xvda', 'Ebs.VolumeSize' => 20 }]
      aws.tags = {
        'Name' => "go-modaemon-dev (Developping by #{ENV['USER']})"
      }
    end

    c.vm.provision :shell, path: File.expand_path('../vagrant/prov/amzn.sh', __FILE__)
    c.vm.synced_folder '.', '/home/ec2-user/src/github.com/mobingi/alm-agent'
  end
end
