#!/bin/bash

set -o errexit

set -o nounset

SUDO=

if [[ -n ${USER:-} && ${UID} -ne 0 ]]; then
  # write a fake user entry with settings matching the host user possible
  echo "${USER}:!:${UID}:${GID}:${HOME}:/bin/bash" >> /etc/passwd
  SUDO='sudo'
fi

${SUDO} service docker start

exec "$@"
