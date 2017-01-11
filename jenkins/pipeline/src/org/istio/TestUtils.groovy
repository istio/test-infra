#!groovy

package org.istio

// Source Code related variables. Set in stashSourceCode.
GIT_SHA = ''
// Global Variables defined in Jenkins
BUCKET = ''
DEFAULT_SLAVE_LABEL = ''
NOTIFY_LIST = ''

// This must be called inside a node
def setGlobals() {
  BUCKET = failIfNullOrEmpty(env.BUCKET, 'BUCKET env must be set.')
  DEFAULT_SLAVE_LABEL = failIfNullOrEmpty(
          env.DEFAULT_SLAVE_LABEL, 'DEFAULT_SLAVE_LABEL env must be set.')
  NOTIFY_LIST = failIfNullOrEmpty(
          env.NOTIFY_LIST, 'NOTIFY_LIST env must be set.')
}

// Updates Pull Request
def updatePullRequest(flow, success = false) {
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
def failIfNullOrEmpty(value, message) {
  if (value == null || value == '') {
    error(message)
  }
  return value
}

// Uses first parameter if set otherwise uses default
def getWithDefault(value, defaultValue = '') {
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
def sendFailureNotification() {
  mail to: "${NOTIFY_LIST}",
      subject: "Job '${env.JOB_NAME}' (${env.BUILD_NUMBER}) failed",
      body: "Please go to ${env.BUILD_URL} to investigate",
      from: 'ESP Jenkins Alerts <esp-alerts-jenkins@google.com>',
      replyTo: 'esp-alerts-jenkins@google.com'
}

// Jenkins use 3 differents slave tyoes
// 1. Normal slave (slave_name): for e2e and other tests
// 2. Build slave (slave_name-build): for all bazel build
// 3. Test slave (slave_name-test): for slave a qualification

// Returns the test slave label from a label
def getTestSlaveLabel(label) {
  return "${label}-test"
}

// Returns the build slave label from a label
def getBuildSlaveLabel(label) {
  return "${label}-build"
}

// Converts a list of [key, value] to a map
def convertToMap(list) {
  def map = [:]
  for (int i = 0; i < list.size(); i++) {
    def key = list.get(i).get(0)
    def value = list.get(i).get(1)
    map[key] = value
  }
  return map
}
// Given a set of pipeline branches, filter according to regex.
def filterBranches(branches, regex) {
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

// In a pipeline, multiple scm checkout might checkout different version of the code.
// Stashing source code to make sure that all pipeline branches uses the same version.
// See JENKINS-35245 bug for more info.
def stashSourceCode(postcheckout_call = null) {
  // Checking out code
  retry(10) {
    // Timeout after 5 minute
    timeout(5) {
      checkout(scm)
    }
    sleep(5)
  }
  if (postcheckout_call != null) {
    postcheckout_call.call()
  }
  // Setting source code related global variable once so it can be reused.
  GIT_SHA = failIfNullOrEmpty(getRevision(), 'GIT_SHA must be set')
  echo('Stashing source code')
  fastStash('src-code', '.')
}

// Checking out code to the current directory
def checkoutSourceCode() {
  deleteDir()
  echo('Unstashing source code')
  fastUnstash('src-code')
  sh("git status")
}

// Checks whether a remote path exists
def pathExistsCloudStorage(filePath) {
  def status = sh(returnStatus: true, script: "gsutil stat ${filePath}")
  return status == 0
}

// Generates the archive path based on the bucket, git_sha and name
def stashArchivePath(name) {
  return "gs://${BUCKET}/${GIT_SHA}/tmp/${name}.tar.gz"
}

// pipeline stash/unstash is too slow as it stores and retrieve data from Jenkins.
// Poor man's stash implementation using Cloud Storage
def fastStash(name, stashPaths) {
  // Checking if archive already exists
  def archivePath = stashArchivePath(name)
  if (!pathExistsCloudStorage(archivePath)) {
    echo("Stashing ${stashPaths} to ${archivePath}")
    retry(5) {
      sh("tar czf - ${stashPaths} | gsutil " +
          "-h Content-Type:application/x-gtar cp - ${archivePath}")
      sleep(5)
    }
  }
}

// Unstashing data to current directory.
def fastUnstash(name) {
  def archivePath = stashArchivePath(name)
  retry(5) {
    sh("gsutil cp ${archivePath} - | tar zxf - ")
    sleep(5)
  }
}

// Finds the revision from the source code.
def getRevision() {
  // Code needs to be checked out for this.
  return sh(returnStdout: true, script: 'git rev-parse --verify HEAD').trim()
}

// Sets an artifacts links to the Build.
def setArtifactsLink() {
  def url = "https://console.cloud.google.com/storage/browser/${BUCKET}/${GIT_SHA}"
  def html = """
<!DOCTYPE HTML>
Find <a href='${url}'>artifacts</a> here
"""
  def artifactsHtml = 'artifacts.html'
  writeFile(file: artifactsHtml, text: html)
  archive(artifactsHtml)
}
