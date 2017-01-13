package org.istio.testutils

import static org.istio.testutils.Utilities.failIfNullOrEmpty

GIT_SHA = ''
BUCKET = ''
NOTIFY_LIST = ''
DEFAULT_SLAVE_LABEL = ''


def initialize() {
  setVars()
  stashSourceCode()
  setArtifactsLink()
}

def setVars() {
  BUCKET = failIfNullOrEmpty(env.BUCKET, 'BUCKET env must be set.')
  DEFAULT_SLAVE_LABEL = failIfNullOrEmpty(
      env.DEFAULT_SLAVE_LABEL, 'DEFAULT_SLAVE_LABEL env must be set.')
  NOTIFY_LIST = failIfNullOrEmpty(env.NOTIFY_LIST, 'NOTIFY_LIST env must be set.')
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
    postcheckout_call()
  }
  // Setting source code related global variable once so it can be reused.
  GIT_SHA = failIfNullOrEmpty(getRevision(), 'Could not find revision')
  echo('Stashing source code')
  fastStash('src-code', '.')
}

// Checks whether a remote path exists
def pathExistsCloudStorage(filePath) {
  def status = sh(returnStatus: true, script: "gsutil stat ${filePath}")
  return status == 0
}

// Checking out code to the current directory
def checkoutSourceCode() {
  deleteDir()
  echo('Unstashing source code')
  fastUnstash('src-code')
  sh("git status")
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

// Finds the revision from the source code.
def getRevision() {
  // Code needs to be checked out for this.
  return sh(returnStdout: true, script: 'git rev-parse --verify HEAD').trim()
}