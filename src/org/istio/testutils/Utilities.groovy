package org.istio.testutils

// Updates Pull Request
def updatePullRequest(flow, success = false) {
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
