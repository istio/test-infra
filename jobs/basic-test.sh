#!/bin/bash

set -e

pwd
ls /workspace
ls /workspace/github.com/nlandolfi
bazel build //...
