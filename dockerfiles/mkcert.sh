#!/bin/sh
/usr/local/bin/mkcert "$@"

chown -R  arcadium:arcadium /etc/certs
chmod go+r /etc/certs/*
ls -la /etc/certs
