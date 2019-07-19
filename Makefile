lint:
	@scripts/run_golangci.sh
	@scripts/check_license.sh
	@bazel run //:buildifier -- -showlog -mode=check $(git ls-files| grep -e BUILD -e WORKSPACE | grep -v vendor)

include Makefile.common.mk
