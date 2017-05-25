#!groovy
@Library('testutils')

import org.istio.testutils.Utilities

// Utilities shared amongst modules
def utils = new Utilities()

node {
  checkout scm
  stage('Cleanup Cluster') {
    utils.initTestingCluster()
    sh('scripts/cleanup-cluster -h 2')
    sleep(60)
  }
}
