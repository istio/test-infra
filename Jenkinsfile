#!groovy

@Library('testutils')

import org.istio.testutils.Utilities
import static org.istio.testutils.Utilities.failIfNullOrEmpty
import org.istio.testutils.GitUtilities

// Slaves Docker tags
UBUNTU_XENIAL = 'ubuntu-16-04'
DOCKER_SLAVES = [
    (UBUNTU_XENIAL): 'gcr.io/istio-testing/ubuntu-16-04-slave'
]

// Source Code related variables. Set in stashSourceCode.
TOOLS_BUCKET = ''

// Utilities shared amongst modules
def gitUtils = new GitUtilities()
def utils = new Utilities()

mainFlow(utils) {
  node {
    gitUtils.initialize()
    TOOLS_BUCKET = failIfNullOrEmpty(env.TOOLS_BUCKET, 'Please set TOOLS_BUCKET env.')
  }
  if (utils.runStage('POSTSUBMIT')) {
    postSubmit(utils)
  }

  if (utils.runStage('STABLE_PRESUBMIT')) {
    slaveUpdate(gitUtils, utils)
  }
}

def postSubmit(utils) {
  node {
    utils.fastForwardStable('istio-testing')
    // Adding extra sleep to prevent Node
    // from ending up in suspended state in Jenkins
    sleep(60)
  }
}

def slaveUpdate(gitUtils, utils) {
  stage('Slave Update') {
    buildNode(gitUtils) {
      nodeLabel = utils.getParam('SLAVE_LABEL', env.DEFAULT_SLAVE_LABEL)
      def dockerImage = "${DOCKER_SLAVES[nodeLabel]}:${env.GIT_SHA}"
      // Test Slave image setup in Jenkins
      def testDockerImage = "${DOCKER_SLAVES[nodeLabel]}:test"
      // Slave image setup in Jenkins
      def finalDockerImage = "${DOCKER_SLAVES[nodeLabel]}:latest"
      echo("Building ${testDockerImage}")
      sh("scripts/jenkins-build-docker-slave -b " +
          "-i ${dockerImage} " +
          "-t ${testDockerImage} " +
          "-s ${nodeLabel} " +
          "-T \"${TOOLS_BUCKET}\"")
      echo("Testing ${testDockerImage}")
      testNode(gitUtils) {
        sh('docker/slaves/slave-test')
      }
      echo("Retagging ${testDockerImage} to ${dockerImage}")
      sh("scripts/jenkins-build-docker-slave " +
          "-i ${testDockerImage} " +
          "-t ${finalDockerImage}")
    }
  }
}
