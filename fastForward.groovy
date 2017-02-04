#!groovy

@Library('testutils')

import org.istio.testutils.Utilities

// Utilities shared amongst modules
def utils = new Utilities()

node {
  def goPath = env.WORKSPACE
  def newWorkspace = "${goPath}/src/main"
  sh("mkdir -p ${newWorkspace}")
  withEnv(["GOPATH=${goPath}", "PATH+GOPATH=${goPath}/bin"]) {
    dir(newWorkspace) {
      stage('Fast Forward') {
        utils.fastForwardStable()
      }
    }
  }
}