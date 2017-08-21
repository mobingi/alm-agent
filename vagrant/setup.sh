#!/bin/bash

mkdir -p /opt/mobingi/alm-agent
wget https://download.labs.mobingi.com/alm-agent/develop/current/alm-agent.tgz
tar xvzf alm-agent.tgz -C /opt/mobingi/alm-agent
ln -s /opt/mobingi/alm-agent/v* /opt/mobingi/alm-agent/current
