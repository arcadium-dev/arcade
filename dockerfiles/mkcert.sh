#!/bin/sh
set -x
umask 0077
/usr/local/bin/mkcert "$@"

chown -R  arcadium:arcadium \
  /etc/certs/client.arcadium.key \
  /etc/certs/assets_key.pem \
  /etc/certs/assets.pem
chmod 0644 /etc/certs/influx* || true
ls -la /etc/certs
