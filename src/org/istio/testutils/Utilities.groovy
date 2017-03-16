package org.istio.testutils

// Updates Pull Request
def updatePullRequest(flow, success = false) {
  if (!getParam('UPDATE_PR', false)) return
  def state, message
  switch (flow) {
    case 'run':
      state = 'PENDING'
      message = "Jenkins job ${env.JOB_NAME} started"
      break
    case 'verify':
      state = success ? 'SUCCESS' : 'FAILURE'
      message = "Jenkins job ${env.JOB_NAME} ${success ? 'passed' : 'failed'}"
      if (success) {
        commentOnPr(message)
      }
      break
    default:
      error('flow can only be run or verify')
  }
  setGitHubPullRequestStatus(
      context: env.JOB_NAME,
      message: message,
      state: state)
}

def runStage(stage) {
  stageToRun = getParam('STAGE', 'ALL')
  if (this.stageToRun == 'ALL') {
    // Stage starting with '_' should be called explicitly
    if (stage.startsWith('_')) {
      return false
    }
    return true
  }
  return stageToRun == stage
}

// Check parameter value and fails if null or empty
static def failIfNullOrEmpty(value, message = 'Value not set') {
  if (value == null || value == '') {
    throw Exception(message)
  }
  return value
}

// Uses first parameter if set otherwise uses default
static def getWithDefault(value, defaultValue = '') {
  if (value == null || value == '') {
    return defaultValue
  }
  return value
}

// Get a build parameter or uses default value.
def getParam(name, defaultValue = '') {
  return getWithDefault(params.get(name), defaultValue)
}

// Creates Token file
def createTokenFile(tokenFile) {
  def credentialId = getParam('GITHUB_TOKEN_ID', env.ISTIO_TESTING_TOKEN_ID)
  withCredentials([string(credentialsId: credentialId, variable: 'GITHUB_TOKEN')]) {
    writeFile(file: tokenFile, text: env.GITHUB_TOKEN)
  }
}

// This Workflow uses the following Build Parameters:
// GITHUB_OWNER: The name of the GitHub org or user, default istio
// GITHUB_REPOS: The name of the repos
// REPO_BASE: The branch to use for base, default stable
// REPO_HEAD: The branch to use for head, default master
// GITHUB_TOKEN_ID: The token id to use in Jenkins. Default is ISTIO_TESTING_TOKEN_ID env variable.
def fastForwardStable() {
  def owner = getParam('GITHUB_OWNER', 'istio')
  def repo = failIfNullOrEmpty(getParam('GITHUB_REPOS'), 'GITHUB_REPOS build parameter needs to be set!')
  def base = getParam('REPO_BASE', 'stable')
  def head = getParam('REPO_HEAD', 'master')
  def tokenFile = '/tmp/token.jenkins'
  createTokenFile(tokenFile)
  sh("github_helper --owner=${owner} " +
      "--repos=${repo} " +
      "--head=${head} " +
      "--base=${base} " +
      "--token_file=${tokenFile} " +
      "--fast_forward " +
      "--verify")
}

def commentOnPr(message) {
  // Passed in by GitHub Integration plugin
  def pr = failIfNullOrEmpty(env.GITHUB_PR_NUMBER)
  def owner = getParam('GITHUB_OWNER', 'istio')
  def repo = failIfNullOrEmpty(getParam('GITHUB_REPO'), 'GITHUB_REPO build parameter needs to be set!')
  def url = "https://api.github.com/repos/${owner}/${repo}/issues/${pr}/comments"
  def credentialId = getParam('GITHUB_TOKEN_ID', env.ISTIO_TESTING_TOKEN_ID)
  withCredentials([string(credentialsId: credentialId, variable: 'GITHUB_TOKEN')]) {
    def curlCommand = "curl -H \"Authorization: token ${GITHUB_TOKEN}\" " +
        "-X POST ${url} --data '{\"body\": \"${message}\"}'"
    sh(curlCommand)
  }
}

// Send Email failure notfication
def sendNotification(notify_list) {
  step([
      $class                  : 'Mailer',
      notifyEveryUnstableBuild: false,
      recipients              : notify_list,
      sendToIndividuals       : true])
}

// init Testing Cluster.
def initTestingCluster() {
  def cluster = failIfNullOrEmpty(env.E2E_CLUSTER, 'E2E_CLUSTER is not set')
  def zone = failIfNullOrEmpty(env.E2E_CLUSTER_ZONE, 'E2E_CLUSTER_ZONE is not set')
  def project = failIfNullOrEmpty(env.PROJECT, 'PROJECT is not set')
  sh('gcloud config set container/use_client_certificate True')
  sh("gcloud container clusters get-credentials " +
      "--project ${project} --zone ${zone} ${cluster}")
}

// Push images to hub
def publishDockerImages(images, tags, config = '') {
  publishDockerImagesToDockerHub(images, tags, config)
  publishDockerImagesToContainerRegistry(images, tags, config)
}

def publishDockerImagesToDockerHub(images, tags, config = '', registry = 'docker.io/istio') {
  withDockerRegistry([credentialsId: env.ISTIO_TESTING_DOCKERHUB]) {
    runReleaseDocker(images, tags, registry, config)
  }
}

def publishDockerImagesToContainerRegistry(images, tags, config = '', registry = 'gcr.io/istio-testing') {
  runReleaseDocker(images, tags, registry, config)
}

def runReleaseDocker(images, tags, registry, config = '') {
  def res = libraryResource('release-docker')
  def releaseDocker = '/tmp/release-docker'
  writeFile(file: releaseDocker, text: res)
  sh("chmod +x ${releaseDocker}")
  sh("${releaseDocker} -h ${registry} -t ${tags} -i ${images} " +
      "${config == '' ? '' : "-c ${config}"}")
}

// Publish Code Coverage
def publishCodeCoverage(credentialId) {
  withCredentials([string(credentialsId: credentialId, variable: 'CODECOV_TOKEN')]) {
    sh('curl -s https://codecov.io/bash | bash /dev/stdin -K')
  }
}

// Converts a list of [key, value] to a map
static def convertToMap(list) {
  def map = [:]
  for (int i = 0; i < list.size(); i++) {
    def key = list.get(i).get(0)
    def value = list.get(i).get(1)
    map[key] = value
  }
  return map
}
// Given a set of pipeline branches, filter according to regex.
static def filterBranches(branches, regex) {
  def filteredBranches = []
  for (int i = 0; i < branches.size(); i++) {
    def keyValue = branches.get(i)
    def key = keyValue.get(0)
    if (key ==~ regex) {
      filteredBranches.add(keyValue)
    }
  }
  return filteredBranches
}

return this
