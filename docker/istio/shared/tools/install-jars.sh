#!/bin/bash

set -eux


JARS_DIR="/usr/local/bin/flakey_jars"
mkdir -p "$JARS_DIR"
wget -q -nc -O /usr/local/bin/flakey_jars/Pair.java https://raw.githubusercontent.com/istio/test-infra/master/flakeyTest/Pair.java
wget -q -nc -O /usr/local/bin/flakey_jars/TotalFlakey.java https://raw.githubusercontent.com/istio/test-infra/master/flakeyTest/TotalFlakey.java
wget -q -nc -O /usr/local/bin/flakey_jars/readPastJunitCommand.sh https://raw.githubusercontent.com/istio/test-infra/master/flakeyTest/readPastJunitCommand.sh
wget -q -nc -O /usr/local/bin/flakey_jars/removeTempFolderCommand.sh https://raw.githubusercontent.com/istio/test-infra/master/flakeyTest/removeTempFolderCommand.sh

wget -q -nc -O jars.zip https://github.com/ChristinaLyu0710/istio-flakey-test/raw/master/FlakeyTest/flexible/Flakey/jars.zip

unzip jars.zip -d "$JARS_DIR"

export PATH=${PATH}:/usr/local/bin