# Sisyphus

### Flakiness Detection and Monitoring Service for Istio

[Design Doc](goo.gl/119VaV)

Sisyphus triggers concurrent reruns on failed e2e tests on the same commit. It extracts JUnit results and integrates with Testgrid to offer test-case level visibility. It records data in persistent stores and visualizes trends by adapting and unifying K8S pipeline -- Kettle, BigQuery Metrics and Velodrome. Email alerts about test failures and flakiness are sent.
