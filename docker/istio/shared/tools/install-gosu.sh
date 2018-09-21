#!/bin/bash

set -eux

GOSU_VERSION="1.4"
GOSU_URL="https://github.com/tianon/gosu/releases/download"
ARCH="$(dpkg --print-architecture)"

gpg --keyserver ha.pool.sks-keyservers.net --recv-keys B42F6819007F00F88E364FD4036A9C25BF357DD4
curl -o /usr/local/bin/gosu -SL "${GOSU_URL}/${GOSU_VERSION}/gosu-${ARCH}"
curl -o /usr/local/bin/gosu.asc -SL "${GOSU_URL}/${GOSU_VERSION}/gosu-${ARCH}.asc"
gpg --verify /usr/local/bin/gosu.asc
rm /usr/local/bin/gosu.asc
chmod +x /usr/local/bin/gosu
