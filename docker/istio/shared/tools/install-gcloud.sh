#!/bin/bash

set -eux

GCLOUD_URL='https://dl.google.com/dl/cloudsdk/channels/rapid/google-cloud-sdk.zip'

export CLOUDSDK_PYTHON_SITEPACKAGES=1
export CLOUDSDK_CORE_DISABLE_PROMPTS=1
export CLOUDSDK_INSTALL_DIR=/usr/lib/
curl https://sdk.cloud.google.com | bash

export PATH="/usr/lib/google-cloud-sdk/bin:${PATH}"

sed -i -e 's/true/false/' /usr/lib/google-cloud-sdk/lib/googlecloudsdk/core/config.json
gcloud -q components update kubectl

