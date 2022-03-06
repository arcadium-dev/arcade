#!/bin/sh
/usr/local/bin/mkcert "$@"

chown -R  arcadium:arcadium /etc/certs
