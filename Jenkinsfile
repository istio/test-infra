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

node {
  gitUtils.initialize()
  TOOLS_BUCKET = failIfNullOrEmpty(env.TOOLS_BUCKET, 'Please set TOOLS_BUCKET env.')
}

node('master') {
  try {
    stage('Slave Update') {
      defaultNode(gitUtils) {
        gitSha = gitUtils.GIT_SHA
        buildNewDockerSlave(gitUtils, utils)
      }
    }
  } catch (Exception e) {
    currentBuild.result = 'FAILURE'
    throw e
  } finally {
    utils.sendNotification(gitUtils.NOTIFY_LIST)
  }
}


def buildNewDockerSlave(gitUtils, utils) {
  nodeLabel = utils.getParam('SLAVE_LABEL', gitUtils.DEFAULT_SLAVE_LABEL)
  def dockerImage = "${DOCKER_SLAVES[nodeLabel]}:${gitUtils.GIT_SHA}"
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
    sh('jenkins/slaves/slave-test')
  }
  echo("Retagging ${testDockerImage} to ${dockerImage}")
  sh("scripts/jenkins-build-docker-slave " +
      "-i ${testDockerImage} " +
      "-t ${finalDockerImage}")
}

