package org.istio.testutils

import static org.istio.testutils.Utilities.getWithDefault


BAZEL_ARGS = ''
BAZEL_BUILD_ARGS = ''


def setVars() {
    BAZEL_ARGS = getWithDefault(env.BAZEL_ARGS)
    BAZEL_BUILD_ARGS = getWithDefault(env.BAZEL_BUILD_ARGS)
}

def fetch(args) {
    timeout(30) {
        retry(3) {
            sh("bazel ${BAZEL_ARGS} fetch ${args}")
            sleep(5)
        }
    }
}

def build(args) {
    timeout(40) {
        retry(3) {
            sh("bazel ${BAZEL_ARGS} build ${BAZEL_BUILD_ARGS} ${args}")
            sleep(5)
        }
    }
}

def test(args) {
    timeout(40) {
        retry(3) {
            sh("bazel ${BAZEL_ARGS} test ${args}")
            sleep(5)
        }
    }
}
