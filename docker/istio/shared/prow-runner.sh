#!/bin/bash

set -eux

service docker start

export HOME=/home/prow

useradd -c "Prow user" -d ${HOME} -u 9001 -G docker,sudo -m prow -s /bin/bash

# Hack for making Pod-utils work on container that have users
PROW_DIRS=( "/logs" "${HOME}" )

for D in "${PROW_DIRS[@]}"; do
  if [[ -d "${D}" ]]; then
    chown -R prow "${D}" || true
  fi
done

[[ -n ${GOPATH:-} ]] && export PATH=${GOPATH}/bin:${PATH}

# Authenticate gcloud, allow failures
if [[ -n "${GOOGLE_APPLICATION_CREDENTIALS:-}" ]]; then
  gcloud auth activate-service-account --key-file="${GOOGLE_APPLICATION_CREDENTIALS}" || true
fi

exec /usr/local/bin/gosu prow "$@"
