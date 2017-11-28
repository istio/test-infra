# GitHub Control

This file triggers release qualification in the latest release pipeline (and automates 0.2.* releases, more information can be found [here](https://github.com/istio/istio/blob/master/release/README.md)).
It is a tool of our own that acts as a GitHub client making REST calls through the GitHub API.

You will need a ```<github token file>``` text file containing the github peronal access token setup following the [instruction](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/)

```
$ git clone https://github.com/istio/test-infra.git
```

and build it using (might need bazel 0.5.4)

```
$ bazel build //toolbox/githubctl
```

The binary output is located in bazel-bin/toolbox/githubctl/githubctl.

```
$ alias githubctl="${PWD}/bazel-bin/toolbox/githubctl/githubctl"
```

To trigger daily release qualification,
```
githubctl --token_file=<github token file> \
	--op=dailyRelQual \
	--hub=<hub of remote docker image registry> \
	--tag=<tag of the release candidate> \
	--gcsPath=<GCS path where istioctl is stored>
```
