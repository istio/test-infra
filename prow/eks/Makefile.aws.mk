# Copyright Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

export KUBECONFIG

# This prevents the kubectl current-context in the execution environment from
# being overwritten unless the intention is made explicit w/ the `save` param.
#
# e.g.
#		make get-cluster-credentials save=true
.PHONY: save-kubeconfig
save-kubeconfig:
ifndef save
	$(eval KUBECONFIG=$(shell mktemp))
endif

# Point kubectl at an EKS cluster. Honors the same `save` convention as the
# gcloud include: without `save=true` a throwaway KUBECONFIG is used so the
# caller's current context is left untouched.
#
# e.g.
#		make get-cluster-credentials save=true

get%cluster-credentials: save-kubeconfig
	aws eks update-kubeconfig --name "$(CLUSTER)" --region "$(REGION)"
