#!/bin/bash

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NGINX_CONF_TEMPLATE="${ROOT}/nginx.conf.template"
NGINX_CONF="${ROOT}/nginx.conf"

# Script Parameters
JENKINS_INSTANCE=''
JENKINS_PORT=''
PROJECT='endpoints-jenkins'
VERSION="$(date +'%Y%m%d%H%M%S')"

while getopts :a:b:c:d:e arg; do
  case ${arg} in
    a) JENKINS_INSTANCE="${OPTARG}";;
    b) JENKINS_PORT="${OPTARG}";;
    c) PROJECT="${OPTARG}";;
    d) VERSION="${OPTARG}";;
    *) echo "Invalid option: -${OPTARG}";;
  esac
done

[[ -n "${JENKINS_INSTANCE}" ]] \
  || { echo 'Please provide Jenkins instance with -a'; exit 1; }
[[ -n "${JENKINS_PORT}" ]] \
  || { echo 'Please provide Jenkins port with -b'; exit 1; }

sed "s|{JENKINS_INSTANCE}|${JENKINS_INSTANCE}|g" "${NGINX_CONF_TEMPLATE}" \
  | sed "s|{JENKINS_PORT}|${JENKINS_PORT}|g" - > "${NGINX_CONF}" \
  || { echo 'Could not create nginx.conf'; exit 1; }

gcloud app deploy \
  --project "${PROJECT}" \
  --version "${VERSION}" "${ROOT}/app.yaml" \
  || { echo 'Could not deploy app'; exit 1; }

URL="https://${VERSION}-dot-${PROJECT}.appspot.com"
echo "Please check ${URL} and promote manually if needed."
