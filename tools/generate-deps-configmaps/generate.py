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

# See README.md for usage information

import os
import sys
import yaml
import argparse
from string import Template


# Get a list of templates in the directory. We remove the suffix here as the name of the
# template is also used in the name of the ConfigMap.
def get_templates(path):
    EXTENSION = ".template"
    templates = []
    for _, _, files in os.walk(path):
        for file in files:
            if file.endswith(EXTENSION):
                templates.append(file.replace(EXTENSION, ""))
    return templates


def str_presenter(dumper, data):
    if len(data.splitlines()) > 1:  # check for multiline string
        return dumper.represent_scalar('tag:yaml.org,2002:str', data, style='|')
    return dumper.represent_scalar('tag:yaml.org,2002:str', data)


def generate_yaml(cm, stream):
    # We need to create a representer to handle multiline strings cleanly and readability
    # otherwise the multiline string would generate quoted with newline characters that do not
    # wrap well in the YAML file and are not easily readable
    yaml.add_representer(str, str_presenter)
    yaml.representer.SafeRepresenter.add_representer(str, str_presenter)
    return yaml.dump_all(cm, stream)


def create_configmap(minor_version, template, namespace="test-pods"):
    template_path = os.path.join(os.path.dirname(
        os.path.realpath(__file__)), template + ".template")

    configmap = {
        "apiVersion": "v1",
        "kind": "ConfigMap",
        "metadata": {
            "name": "release-" + minor_version + "-" + template,
            "namespace": namespace
        },
        "data": {
            "dependencies": ""
        }
    }

    # Load the template
    f = open(template_path, "r")
    template = Template(f.read())
    f.close()

    configmap["data"]["dependencies"] = template.safe_substitute(
        minor_version=minor_version,
    )

    return configmap


def load_configmaps(path):
    f = open(path, "r")
    # Convert the Generator class to a list so that we can append the new ConfigMaps to it
    release_deps = list(yaml.safe_load_all(f))
    return release_deps


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--destination",
        # Default directory is relative to script, ../../prow/cluster/private/release-deps.yaml
        default=os.path.join(
            os.path.dirname(os.path.realpath(__file__)),
            "../../prow/cluster/private/release-deps.yaml"
        ),
        help="Path to the release-deps.yaml file",
    )
    parser.add_argument(
        "--templates",
        default=os.path.dirname(os.path.realpath(__file__)),
        help="Path to the directory containing the templates",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
    )
    parser.add_argument(
        "--minor-version",
        required=True,
        help="Minor version of the release",
    )
    parser.add_argument(
        "--namespace",
        default="test-pods",
        help="Namespace to deploy the newly generated ConfigMaps",
    )
    args = parser.parse_args()

    if not os.path.exists(args.templates) or not os.path.isdir(args.templates):
        print("The specified template location " + args.templates +
              " does not exist or is not a directory")
        sys.exit(1)

    if not os.path.exists(args.destination) or not os.path.isfile(args.destination):
        print("The specified destination " + args.destination +
              " does not exist or is not a file")
        sys.exit(1)

    # This is primarily done so we can cleanly append the new ConfigMaps to the existing
    release_deps = load_configmaps(args.destination)

    templates = get_templates(args.templates)

    for template in templates:
        print("Generating ConfigMap for " + template)
        cm = create_configmap(args.minor_version, template, args.namespace)
        release_deps.append(cm)

    if not args.dry_run:
        print("Writing to " + args.destination)
        with open(args.destination, "w") as stream:
            stream.write(
                "# Below contains the configmaps that configure the release-builder manifests in istio/istio and istio/release-builder.\n"
            )
            stream.write("# This should be updated for each version\n")
            generate_yaml(release_deps, stream)
    else:
        print("")
        generate_yaml(release_deps, sys.stdout)


if __name__ == "__main__":
    main()
