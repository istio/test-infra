#!/bin/bash

set -e

echo "PRESENT WORKING DIR:"
pwd
bazel build //...
