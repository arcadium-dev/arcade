#!/bin/sh
#set -x
umask 0077
/usr/local/bin/mkcert "$@"

chown -R arcadium:arcadium \
  /etc/certs/client.arcadium.key \
  /etc/certs/assets_key.pem \
  /etc/certs/assets.pem \
  /etc/certs/rootCA.pem
chmod 0644 /etc/certs/influx* || true

cp /etc/certs/rootCA.pem /etc/certs/postgres_rootCA.pem
chown -R 70:70 \
  /etc/certs/postgres.pem \
  /etc/certs/postgres_key.pem \
  /etc/certs/postgres_rootCA.pem
chmod 0600 /etc/certs/postgres* || true
