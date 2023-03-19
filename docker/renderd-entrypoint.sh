#!/usr/bin/env sh
set -eux

if [ ! -f /data/database/planet-import-complete ]; then
  /run.sh import
fi

sed --in-place \
  --expression "s/socketname=.*/iphostname=0.0.0.0\nipport=7654/g" \
  /etc/renderd.conf

/run.sh run
