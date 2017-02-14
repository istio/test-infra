#!groovy

@Library('testutils')

import org.istio.testutils.Utilities

// Utilities shared amongst modules
def utils = new Utilities()

node {
  stage('Fast Forward') {
    utils.fastForwardStable()
  }
}