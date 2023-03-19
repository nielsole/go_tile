#!/usr/bin/env sh
set -eux

./server \
  --data /data/tiles \
  --map default \
  --socket renderd:7654 \
  --static ./static
