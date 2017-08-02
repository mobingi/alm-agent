#!/bin/bash

INFO_FILE="./bin/version_info.json"

if ! [ -f ${INFO_FILE} ] ; then
  echo 'JSONfile not found.'
  exit 1
fi

AWSCLI="/usr/local/bin/aws"
bucket="download.labs.mobingi.com"
version=`cat ${INFO_FILE} | jq -r .version`

mv bin ${version}
tar cvzf ${version}/go-modaemon.tgz ${version}/*
$AWSCLI s3 cp --recursive ${version}/ s3://${bucket}/go-modaemon/${CIRCLE_BRANCH}/${version}/
$AWSCLI s3 cp --recursive ${version}/ s3://${bucket}/go-modaemon/${CIRCLE_BRANCH}/current/
