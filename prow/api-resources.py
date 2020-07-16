#!/usr/bin/env python3

# Copyright 2020 Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from argparse import ArgumentParser
from collections import defaultdict
from itertools import accumulate
from re import findall
from subprocess import run, PIPE, CalledProcessError


def fetch_resources():
    try:
        r = run(
            ["kubectl", "api-resources"], check=True, stdout=PIPE, encoding="utf-8"
        ).stdout
        r = r.split("\n")[:-1]

        head, body = r[0], r[1:]

        col_w = list(
            accumulate([0] + [len(c) for c in findall(r"\w+\s+", head)] + [len(head)])
        )
        row = [[] for _ in range(len(body))]

        for i, line in enumerate(body):
            for j in range(len(col_w) - 1):
                row[i].append(line[col_w[j]: col_w[j + 1]].strip())

        return row

    except CalledProcessError:
        return []


def fetch_version(name, version):
    try:
        run(
            ["kubectl", "explain", name, "--api-version", version],
            check=True,
            stdout=PIPE,
            stderr=PIPE,
        )

        return True

    except CalledProcessError:
        return False


def fetch_versions():
    try:
        v = run(
            ["kubectl", "api-versions"], check=True, stdout=PIPE, encoding="utf-8"
        ).stdout
        v = v.split("\n")[:-1]

        versions_by_group = defaultdict(list)

        for gv in v:
            group, _, version = gv.rpartition("/")
            versions_by_group[group].append(version)

        return versions_by_group

    except CalledProcessError:
        return {}


def main(delimiter, format, group_denylist, kind_denylist):
    resources = fetch_resources()
    versions = fetch_versions()
    group_denylist = set(group_denylist)
    kind_denylist = set(kind_denylist)
    r = []

    for name, _, group, _, kind in resources:
        if kind in kind_denylist or group in group_denylist:
            continue

        for version in versions.get(group, []):
            if fetch_version(name, f"{group}/{version}"):
                r.append(f"{group or 'core'}/{version}/{kind}")

    return delimiter.join([format % v for v in r])


if __name__ == "__main__":
    parser = ArgumentParser()
    parser.add_argument(
        "--delimiter",
        type=str,
        default="\n",
        help="delimiter string for the list of api resources.",
    )
    parser.add_argument(
        "--format", type=str, default="%s", help="format string for each api resource."
    )
    parser.add_argument(
        "--group-denylist",
        type=str,
        nargs="*",
        default=["authentication.k8s.io", "authorization.k8s.io"],
        help="set of api groups to denylist.",
    )
    parser.add_argument(
        "--kind-denylist",
        type=str,
        nargs="*",
        default=["Binding"],
        help="set of api kinds to denylist.",
    )
    args = parser.parse_args()
    print(main(args.delimiter, args.format, args.group_denylist, args.kind_denylist))
