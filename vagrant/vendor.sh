#!/bin/bash

set -xe

mkdir -p /tmp/vendor /tmp/vendor.orig

if [ ! -d `pwd`/vendor ] ; then
  mkdir `pwd`/vendor
fi

if [ ! -d `pwd`/vendor.orig ] ; then
  mkdir `pwd`/vendor.orig
fi

mount -o bind /tmp/vendor `pwd`/vendor
mount -o bind /tmp/vendor.orig `pwd`/vendor.orig
