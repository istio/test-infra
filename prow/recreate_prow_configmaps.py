#!/usr/bin/env python

# Copyright 2018 The Kubernetes Authors.
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

#
# USAGE: have KUBECONFIG pointed at your prow cluster then from test-infra root:
#
# prow/recreate_prow_configmaps.py [--wet]
#

from __future__ import print_function

import os
import sys

import argparse
import subprocess


def recreate_prow_config(wet, configmap_name, path):
    print('recreating prow config:')
    cmd = (
        'kubectl create configmap %s'
        ' --from-file=config.yaml=%s'
        ' --dry-run -o yaml | kubectl replace configmap %s -f -'
    ) % (configmap_name, path, configmap_name)
    real_cmd = ['/bin/sh', '-c', cmd]
    print(real_cmd)
    if wet:
        subprocess.check_call(real_cmd)


def recreate_plugins_config(wet, configmap_name, path):
    print('recreating plugins config:')
    cmd = (
        'kubectl create configmap %s'
        ' --from-file=plugins.yaml=%s'
        ' --dry-run -o yaml | kubectl replace configmap %s -f -'
    ) % (configmap_name, path, configmap_name)
    real_cmd = ['/bin/sh', '-c', cmd]
    print(real_cmd)
    if wet:
        subprocess.check_call(real_cmd)


def recreate_job_config(wet, job_configmap, job_config_dir):
    print('recreating jobs config:')

    # regenerate
    paths = []
    cmd = ["kubectl", "create", "configmap", job_configmap]
    for root, _, files in os.walk(job_config_dir):
        for name in files:
            if name.endswith(".yaml"):
                path = os.path.join(root, name)
                real_cmd = ['/bin/sh', '-c', 'gzip -k ' + path]
                print(real_cmd)
                if wet:
                    subprocess.check_call(real_cmd)
                paths.append(path)
                cmd.append('--from-file=%s=%s' % (name, path + '.gz'))
    cmd.append('--dry-run -o yaml | kubectl replace configmap %s -f -' % (job_configmap))
    real_cmd = ['/bin/sh', '-c', ' '.join(cmd)]
    print(real_cmd)
    if wet:
        subprocess.check_call(real_cmd)
    for path in paths:
        real_cmd = ['/bin/sh', '-c', 'rm ' + path + '.gz']
        print(real_cmd)
        if wet:
            subprocess.check_call(real_cmd)


def main():
    parser = argparse.ArgumentParser()
    # jobs config
    parser.add_argument("--job-configmap", default="job-config", help="name of prow jobs configmap")
    parser.add_argument(
        "--job-config-dir", default="config/jobs",
        help="root dir of prow jobs configmap")
    # prow config
    parser.add_argument("--prow-configmap", default="config",
                        help="name of prow primary configmap")
    parser.add_argument(
        "--prow-config-path", default="config.yaml",
        help="path to the primary prow config")
    # plugins config
    parser.add_argument("--plugins-configmap", default="plugins",
                        help="name of prow plugins configmap")
    parser.add_argument(
        "--plugins-config-path", default="plugins.yaml",
        help="path to the prow plugins config")
    # wet or dry?
    parser.add_argument("--wet", action="store_true")
    parser.add_argument("--silent", action="store_true",
                        help="if confirmation is needed for the change")
    args = parser.parse_args()

    # debug the current context
    out = subprocess.check_output(['kubectl', 'config', 'current-context'])
    print('Current KUBECONFIG context: ' + out)

    # require additional confirmation in --wet mode
    prompt = '!' * 65 + (
        "\n!!     WARNING THIS WILL RECREATE **ALL** PROW CONFIGMAPS.     !!"
        "\n!!    ARE YOU SURE YOU WANT TO DO THIS? IF SO, ENTER 'YES'.    !! "
    ) + '\n' + '!' * 65 + '\n\n: '
    if args.wet and not args.silent:
        if raw_input(prompt) != "YES":
            print("you did not enter 'YES'")
            sys.exit(-1)

    # first prow config
    recreate_prow_config(args.wet, args.prow_configmap, args.prow_config_path)
    print('')
    # then plugins config
    recreate_plugins_config(args.wet, args.plugins_configmap, args.plugins_config_path)
    print('')
    # finally jobs config
    recreate_job_config(args.wet, args.job_configmap, args.job_config_dir)


if __name__ == '__main__':
    main()
