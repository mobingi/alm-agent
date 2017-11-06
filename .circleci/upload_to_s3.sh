#!/bin/bash

INFO_FILE="version_info.json"
INFO_PATH="./bin/${INFO_FILE}"

if ! [ -f ${INFO_PATH} ] ; then
  echo 'INFO_FILE not found.'
  exit 1
fi

AWSCLI="/usr/local/bin/aws"
bucket="download.labs.mobingi.com"
version=`cat ${INFO_PATH} | jq -r .version`

mv bin ${version}
tar cvzf ${version}/alm-agent.tgz ${version}/*

$AWSCLI s3 cp ${version}/alm-agent.tgz s3://${bucket}/alm-agent/${CIRCLE_BRANCH}/${version}/
$AWSCLI s3 cp --cache-control 'max-age=3600' ${version}/alm-agent.tgz s3://${bucket}/alm-agent/${CIRCLE_BRANCH}/current/

$AWSCLI s3 cp --content-type text/json ${version}/${INFO_FILE} s3://${bucket}/alm-agent/${CIRCLE_BRANCH}/${version}/
$AWSCLI s3 cp --content-type text/json --cache-control 'max-age=1800' ${version}/${INFO_FILE} s3://${bucket}/alm-agent/${CIRCLE_BRANCH}/current/
