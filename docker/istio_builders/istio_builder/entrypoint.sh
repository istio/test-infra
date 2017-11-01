#!/bin/bash

set -o errexit

set -o nounset

# write a fake user entry with settings matching the host user possible
echo "${USER}:!:${UID}:${GID}:${HOME}:/bin/bash" >> /etc/passwd

exec "$@"
