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

# upload for use guess per file
for x in go-modaemon.tgz $INFO_FILE ; do
  $AWSCLI s3 cp ${version}/${x} s3://${bucket}/go-modaemon/${CIRCLE_BRANCH}/${version}/${x}
  $AWSCLI s3 cp ${version}/${x} s3://${bucket}/go-modaemon/${CIRCLE_BRANCH}/current/${x}
done