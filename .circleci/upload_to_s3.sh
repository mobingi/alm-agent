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
tar cvzf ${version}/go-modaemon.tgz ${version}/*

$AWSCLI s3 cp ${version}/go-modaemon.tgz s3://${bucket}/go-modaemon/${CIRCLE_BRANCH}/${version}/
$AWSCLI s3 cp ${version}/go-modaemon.tgz s3://${bucket}/go-modaemon/${CIRCLE_BRANCH}/current/

$AWSCLI s3 cp --content-type application/json ${version}/${INFO_FILE} s3://${bucket}/go-modaemon/${CIRCLE_BRANCH}/${version}/
$AWSCLI s3 cp --content-type application/json ${version}/${INFO_FILE} s3://${bucket}/go-modaemon/${CIRCLE_BRANCH}/current/
