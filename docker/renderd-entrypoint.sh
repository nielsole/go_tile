#!/usr/bin/env bash
set -euxo pipefail

if [ ! -f /data/database/planet-import-complete ]; then
  /run.sh import
fi

/usr/bin/sed --in-place \
  --expression "s/socketname=.*/iphostname=0.0.0.0\nipport=7654/g" \
  /etc/renderd.conf

/run.sh run
