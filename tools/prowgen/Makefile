# Copyright Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

PROJECT = istio-testing
HUB = gcr.io
TAG = $(shell date +v%Y%m%d)-$(shell git describe --always --dirty --match '^$$')

.PHONY: deploy
deploy: configure-docker image push

.PHONY: activate-serviceaccount
activate-serviceaccount:
ifdef GOOGLE_APPLICATION_CREDENTIALS
	gcloud auth activate-service-account --key-file="$(GOOGLE_APPLICATION_CREDENTIALS)"
endif

.PHONY: configure-docker
configure-docker: activate-serviceaccount
	gcloud auth configure-docker

.PHONY: image
image:
	docker build -t "$(HUB)/$(PROJECT)/prowgen:$(TAG)" -t "$(HUB)/$(PROJECT)/prowgen:latest" -f Dockerfile ../../

.PHONY: push
push:
	docker push "$(HUB)/$(PROJECT)/prowgen:$(TAG)"
	docker push "$(HUB)/$(PROJECT)/prowgen:latest"
