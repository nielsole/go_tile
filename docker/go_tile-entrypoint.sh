#!/usr/bin/env sh
set -eux

OPTIONS="-static ${STATIC_DIR:-/usr/share/go_tile/static}"

if [ -n "${MAP_NAME:-}" ]; then
  OPTIONS="${OPTIONS:-} -map ${MAP_NAME}"
fi

if [ -n "${METATILE_DIR:-}" ]; then
  OPTIONS="${OPTIONS:-} -data ${METATILE_DIR}"
fi

if [ -n "${RENDERD_SOCKET:-}" ]; then
  OPTIONS="${OPTIONS:-} -socket ${RENDERD_SOCKET}"
fi

go_tile ${OPTIONS}
