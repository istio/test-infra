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

main(utils) {
  if (utils.runStage('_SLAVE_UPDATE')) {
    slaveUpdate(gitUtils, utils)
  }
  if (utils.runStage('_FAST_FORWARD')) {
    fastForwardStable('istio-testing')
  }
}

def slaveUpdate(gitUtils, utils) {
  stage('Slave Update') {
    defaultNode(gitUtils) {
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
        sh('docker/slaves/slave-test')
      }
      echo("Retagging ${testDockerImage} to ${dockerImage}")
      sh("scripts/jenkins-build-docker-slave " +
          "-i ${testDockerImage} " +
          "-t ${finalDockerImage}")
    }
  }
}

def fastForwardStable(repo, base = 'stable', head = 'master', owner = 'istio') {
  goBuildNode(gitUtils, 'main') {
    stage('Fast Forward') {
      def githubPr = libraryResource('github_pr.go')
      def tokenFile = '/tmp/gh.token'
      def credentialId = env.ISTIO_TESTING_TOKEN_ID
      withCredentials([string(credentialsId: credentialId, variable: 'GITHUB_TOKEN')]) {
        writeFile(file: tokenFile, env.GITHUB_TOKEN)
      }
      writeFile(file: 'gh.go', text: githubPr)
      sh "go get ./..."
      sh "go run gh.go --owner=${owner} " +
          "--repo=${repo} " +
          "--head=${head} " +
          "--base=${base} " +
          "--token_file=${tokenFile} " +
          "--fast_forward " +
          "--verify"
    }
  }
}


