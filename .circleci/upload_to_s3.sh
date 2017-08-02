#!/bin/bash

AWSCLI="/usr/local/bin/aws"
bucket="download.labs.mobingi.com"
version=`cat ${CIRCLE_BRANCH}.json | jq -r .version`

$AWSCLI s3 cp ${CIRCLE_BRANCH}.json s3://${bucket}/go-modaemon/
$AWSCLI s3 cp bin/* s3://${bucket}/go-modaemon/${CIRCLE_BRANCH}/${version}/
$AWSCLI s3 cp bin/* s3://${bucket}/go-modaemon/${CIRCLE_BRANCH}/current/
