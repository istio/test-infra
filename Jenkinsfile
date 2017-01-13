#!groovy

@Library('testutils')

def utils = new org.istio.TestUtils()

// Supported VM Images
UBUNTU_XENIAL = 'ubuntu-16-04'

// Slaves Docker tags
DOCKER_SLAVES = [
    (UBUNTU_XENIAL): 'gcr.io/endpoints-jenkins/ubuntu-16-04-slave'
]

// Source Code related variables. Set in stashSourceCode.
TOOLS_BUCKET = ''

node {
  utils.setGlobals()
  utils.stashSourceCode()
  utils.setArtifactsLink()
  TOOLS_BUCKET = utils.failIfNullOrEmpty(env.TOOLS_BUCKET, 'Please set TOOLS_BUCKET env.')
}

node('master') {
  def nodeLabel = utils.getParam('SLAVE_LABEL', utils.DEFAULT_SLAVE_LABEL)
  try {
    stage('Slave Update') {
      node(nodeLabel) {
        buildNewDockerSlave(nodeLabel)
      }
    }
  } catch (Exception e) {
    currentBuild.result = 'FAILURE'
    throw e
  } finally {

    step([
        $class                  : 'Mailer',
        notifyEveryUnstableBuild: false,
        recipients              : utils.NOTIFY_LIST,
        sendToIndividuals       : true])
  }
}

def buildNewDockerSlave(nodeLabel) {
  utils.checkoutSourceCode()
  def dockerImage = "${DOCKER_SLAVES[nodeLabel]}:${utils.GIT_SHA}"
  // Test Slave image setup in Jenkins
  def testDockerImage = "${DOCKER_SLAVES[nodeLabel]}:test"
  // Slave image setup in Jenkins
  def finalDockerImage = "${DOCKER_SLAVES[nodeLabel]}:latest"
  echo("Building ${testDockerImage}")
  sh("script/jenkins-build-docker-slave -b " +
      "-i ${dockerImage} " +
      "-t ${testDockerImage} " +
      "-s ${nodeLabel} " +
      "-T \"${TOOLS_BUCKET}\"")
  echo("Testing ${testDockerImage}")
  node(utils.getTestSlaveLabel(nodeLabel)) {
    utils.checkoutSourceCode()
    sh('jenkins/slaves/slave-test')
  }
  echo("Retagging ${testDockerImage} to ${dockerImage}")
  sh("script/jenkins-build-docker-slave " +
      "-i ${testDockerImage} " +
      "-t ${finalDockerImage}")
}
