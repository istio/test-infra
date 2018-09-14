#!/bin/bash

set -o errexit
set -o nounset

service docker start

export HOME=/home/prow

useradd -c "Prow user" -d ${HOME} -u 9001 -G docker,sudo -m prow -s /bin/bash

# Hack for making Pod-utils work on container that have users
PROW_DIRS=( "/logs" "${HOME}" )

for D in "${PROW_DIRS[@]}"; do
  if [[ -d "${D}" ]]; then
    chown -R prow "${D}"
  fi
done

exec /usr/local/bin/gosu prow "$@"
