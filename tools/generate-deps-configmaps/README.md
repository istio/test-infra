# generate-deps-ConfigMaps

Used to generate the release-deps and istio-deps ConfigMaps for the `istio-private` Prow environment. This is needed for the security release processes
to ensure the jobs are depending on the appropriate repositories in the `istio-private` org.

This should ideally be run during Step 3 of the branch cutting process, see: [istio#46985](https://github.com/istio/istio/issues/46985).
This should be run during the steps of Step 3 where the release branch jobs are being generated, though specific order does not matter.

## Prerequisites

- [Python 3.6+](https://www.python.org/downloads/)
- [PyYAML](https://pyyaml.org/wiki/PyYAMLDocumentation)

## Templates

The templates utilize [Template strings](https://docs.python.org/3.4/library/string.html#template-strings) to generate the ConfigMap dependency data.

## Usage

```bash
python3 tools/generate-deps-ConfigMaps/generate.py [options]...
```

You can run this within `make shell` to ensure you have the appropriate Python and PyYAML versions.

### Options

The following is a list of supported options. If an option is optional, then its _default_ value will be used if not provided. If an option is required
and not provided, the execution will fail with a non-zero exit code.

| Option | Required | Default | Description |
| ------ | -------- | ------- | ----------- |
| `--destination` | No | from root of repo: `prow/cluster/private/release-deps.yaml` | The path to the release-deps ConfigMap. New ConfigMaps will be appended and this assumes it is a valid YAML file. |
| `--dry-run` | No | false | If set, the script will not write the ConfigMaps to disk, but dump the generated YAML to stdout. |
| `--minor-version` | Yes | N/A | The minor version of the release. |
| `--namespace` | No | `test-pods` | Namespace to apply the new ConfigMaps to. This does not change existing ConfigMaps in the destination file. |
| `--templates` | No | script's directory | Path to directory containing the templates to use. |
