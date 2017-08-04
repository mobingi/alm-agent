#!/bin/bash

mkdir -p /opt/mobingi/go-modaemon
wget https://download.labs.mobingi.com/go-modaemon/develop/current/go-modaemon.tgz
tar xvzf go-modaemon.tgz -C /opt/mobingi/go-modaemon
ln -s /opt/mobingi/go-modaemon/v* /opt/mobingi/go-modaemon/current
