#!/bin/bash

mkdir -p /opt/mobingi/alm-agent /opt/mobingi/etc
wget https://download.labs.mobingi.com/alm-agent/develop/current/alm-agent.tgz
tar xvzf alm-agent.tgz -C /opt/mobingi/alm-agent
ln -sf /opt/mobingi/alm-agent/v* /opt/mobingi/alm-agent/current
