#!/bin/bash
echo "# Run update-prow-staging.sh to keep sync this with the master config" > istio.istio.prow-staging.yaml
cat istio.istio.master.yaml >> istio.istio.prow-staging.yaml
sed -i 's/-master/-prow-staging/g' istio.istio.prow-staging.yaml
sed -i 's/\^master/\^prow-staging/g' istio.istio.prow-staging.yaml