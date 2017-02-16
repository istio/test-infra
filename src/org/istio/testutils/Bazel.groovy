package org.istio.testutils

BAZEL_BUILD_ARGS = ''
BAZEL_RUN_ARGS = ''
BAZEL_STARTUP_ARGS = ''
BAZEL_TEST_ARGS = ''

def defaultValue(first, second) {
  def value = first == null ? second : first
  return value == null ? '' : value
}

def setVars(startup = null, build = null, test = null, run=null) {
  BAZEL_BUILD_ARGS = defaultValue(build, env.BAZEL_BUILD_ARGS)
  BAZEL_RUN_ARGS = defaultValue(run, env.BAZEL_RUN_ARGS)
  BAZEL_STARTUP_ARGS = defaultValue(startup, env.BAZEL_STARTUP_ARGS)
  BAZEL_TEST_ARGS = defaultValue(test, env.BAZEL_TEST_ARGS)
}

def fetch(args) {
  timeout(30) {
    retry(3) {
      sh("bazel ${BAZEL_STARTUP_ARGS} fetch ${args}")
      sleep(5)
    }
  }
}

def build(args) {
  timeout(40) {
    retry(3) {
      sh("bazel ${BAZEL_STARTUP_ARGS} build ${BAZEL_BUILD_ARGS} ${args}")
      sleep(5)
    }
  }
}

def test(args) {
  timeout(40) {
    sh("bazel ${BAZEL_STARTUP_ARGS} test ${BAZEL_TEST_ARGS} ${args}")
    sleep(5)
  }
}

def run(args) {
  sh("bazel ${BAZEL_STARTUP_ARGS} run ${BAZEL_RUN_ARGS} ${args}")
  sleep(5)
}

def version() {
  sh('bazel version')
}

def updateBazelRc(updateBazelrc = '.bazelrc.jenkins') {
  if (fileExists(updateBazelrc)) {
    sh("cat ${updateBazelrc} >> .bazelrc")
  }
}

return this
