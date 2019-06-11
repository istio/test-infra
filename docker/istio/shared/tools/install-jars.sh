#!/bin/bash

set -eux

JARS_DIR="/usr/local/bin/flakey_jars"
wget -q -nc -O jars.zip https://github.com/ChristinaLyu0710/istio-flakey-test/raw/master/FlakeyTest/flexible/Flakey/jars.zip
mkdir -p "$JARS_DIR"
unzip jars.zip -d "$JARS_DIR"

export PATH=${PATH}:/usr/local/bin 