This directory contains standard images maintained by Istio for test infrastructure purposes.

* prowbazel - gives an environment with git, gcloud, python, bazel, and go.
    It is suitable for bazel-based projects. Currently, it is analagous to the
    Jenkins slave image, and therefore can be used by all jobs for testing
    istio repositories. In the future, specific jobs may have specific images.
* bazel-remote & slaves - provide images for Jenkins-based testing.
