#!/bin/bash

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

# The file depends on the java application,
# jars dependency folder and java files from flakeyTest folder.
# The commands in the file runs the java files from flakeyTest
# from prow daily to calculate the percentage of flakey-ness of
# test cases.
cd flakeyTest

/usr/local/bin/java/bin/javac -cp ".:/usr/local/bin/flakey_jars/jars/*" Pair.java TotalFlakey.java

/usr/local/bin/java/bin/java -cp ".:/usr/local/bin/flakey_jars/jars/*" TotalFlakey