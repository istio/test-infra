#!/bin/bash

set -o errexit

set -o nounset

if [[ -n ${USER:-} ]]; then
  # write a fake user entry with settings matching the host user possible
  echo "${USER}:!:${UID}:${GID}:${HOME}:/bin/bash" >> /etc/passwd
fi

exec "$@"
