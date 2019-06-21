#!/bin/bash

set -eux


YAMLLINTL_BASE_URL='http://launchpadlibrarian.net/411563953/yamllint_1.15.0-1_all.deb'
YAMLLINTL_DEB="yamllint_1.15.0-1_all.deb"

wget -q -nc "${YAMLLINTL_BASE_URL}"
chmod +x "${YAMLLINTL_DEB}"
apt-get update
apt-get -qqy install python3-pkg-resources python3-yaml python3-pathspec 
dpkg -i "${YAMLLINTL_DEB}"

