package org.istio.testutils

// Updates Pull Request
def updatePullRequest(flow, success = false) {
  if (!getParam('UPDATE_PR', false)) return
  def state, message
  switch (flow) {
    case 'run':
      state = 'PENDING'
      message = "Running presubmits at ${env.BUILD_URL} ..."
      break
    case 'verify':
      state = success ? 'SUCCESS' : 'FAILURE'
      message = "${success ? 'Successful' : 'Failed'} presubmits. " +
          "Details at ${env.BUILD_URL}."
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
static def failIfNullOrEmpty(value, message) {
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

// This Workflow uses the following Build Parameters:
// FF_OWNER: The name of the Githug org or user, default istio
// FF_REPO: The name of the repo
// FF_BASE: The branch to use for base, default stable
// FF_HEAD: The branch to use for head, default master
// FF_TOKEN_ID: The token id to use in Jenkins. Default is set to ISTIO_TESTING_TOKEN_ID env variable.
def fastForwardStable() {
  def res = libraryResource('github_pr.go')
  def tokenFile = '/tmp/gh.token'
  def owner = getParam('FF_OWNER', 'istio')
  def repo = failIfNullOrEmpty(getParam('FF_REPO'), 'FF_REPO build parameter needs to be set!')
  def base = getParam('FF_BASE', 'stable')
  def head = getParam('FF_HEAD', 'master')
  def credentialId = getParam('FF_TOKEN_ID', env.ISTIO_TESTING_TOKEN_ID)
  withCredentials([string(credentialsId: credentialId, variable: 'GITHUB_TOKEN')]) {
    writeFile(file: tokenFile, text: env.GITHUB_TOKEN)
  }
  writeFile(file: 'gh.go', text: res)
  sh("go get ./...")
  sh("go run gh.go --owner=${owner} " +
      "--repo=${repo} " +
      "--head=${head} " +
      "--base=${base} " +
      "--token_file=${tokenFile} " +
      "--fast_forward " +
      "--verify")
}

// Send Email failure notfication
def sendNotification(notify_list) {
  step([
      $class                  : 'Mailer',
      notifyEveryUnstableBuild: false,
      recipients              : notify_list,
      sendToIndividuals       : true])
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