#!/usr/bin/env sh
set -eux

go_tile \
  --data /data/tiles \
  --map default \
  --socket renderd:7654 \
  --static /usr/share/go_tile/static
