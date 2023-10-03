#!/bin/sh
set -x
umask 0077
/usr/local/bin/mkcert "$@"

#chown -R  arcadium:arcadium /etc/certs
chmod 0644 /etc/certs/influx* || true
ls -la /etc/certs
