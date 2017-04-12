package org.istio.testutils

// Get a build parameter or use default value.
def getParam(name, defaultValue = '') {
  def value = params.get(name)
  if (value == null || value == '') {
    return defaultValue
  }
  return value
}

def initialize(Closure postcheckoutCall = null) {
  stashSourceCode(postcheckoutCall)
  setArtifactsLink()
}

def setGit() {
  writeFile(
      file: "${env.HOME}/.gitconfig",
      text: '''
[user]
        name = istio-testing
        email = istio-testing@gmail.com
[remote "origin"]
        fetch = +refs/pull/*/head:refs/remotes/origin/pr/*''')
}

// In a pipeline, multiple scm checkout might checkout different version of the code.
// Stashing source code to make sure that all pipeline branches uses the same version.
// See JENKINS-35245 bug for more info.
def stashSourceCode(Closure postcheckoutCall = null) {
  // Checking out code
  retry(10) {
    // Timeout after 5 minute
    timeout(5) {
      checkout(scm)
    }
    sleep(5)
  }
  updateSubmodules()
  if (postcheckoutCall != null) {
    postcheckoutCall()
  }
  // Setting source code related global variable once so it can be reused.
  def gitCommit = getRevision()
  env.GIT_SHA = gitCommit
  env.GIT_COMMIT = gitCommit
  env.GIT_SHORT_SHA = gitCommit.take(7)
  env.GIT_TAG = getTag()
  env.GIT_BRANCH = getBranch()
  // Might not be safe to share envs.
  sh('env')

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
  setGit()
}

// Get the reference to use
def getRef() {
  def useTag = getParam('USE_TAG', false)
  def ref = env.GIT_SHA
  if (useTag && env.GIT_TAG != '') {
    ref = env.GIT_TAG
  }
  return ref
}

// Base Path
def basePath(name = '') {
  def ref = getRef()
  def path = "gs://${env.BUCKET}/${ref}"
  if (name != '') {
    path = "${path}/${name}"
  }
  return path
}

// Artifacts Path
def artifactsPath(name = '') {
  def path = basePath('artifacts')
  if (name != '') {
    path = "${path}/${name}"
  }
  return path
}

// Generates the archive path based on the bucket, git_sha and name
def stashArchivePath(name) {
  return basePath("tmp/${name}.tar.gz")
}

def logsPath() {
  return basePath("logs")
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

// This only trigger when SUBMODULES_UPDATE build parameter is set.
// Format is FILE1:KEY1:VALUE1,FILE2:KEY2:VALUE2
// Which will update the key1 in file1 with the new value
// and key2 in file2 with the value2 and create a commit for each change
def updateSubmodules() {
  def submodules_update = getParam('SUBMODULES_UPDATE', false)
  if (!submodules_update) {
    return
  }
  def res = libraryResource('update-submodules')
  def remoteFile = '/tmp/update-submodules'
  writeFile(file: remoteFile, text: res)
  sh("chmod +x ${remoteFile}")
  sh("${remoteFile} -s ${submodules_update}")
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
  def ref = getRef()
  def url = "https://console.cloud.google.com/storage/browser/${env.BUCKET}/${ref}"
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

def getTag() {
  return sh(returnStdout: true, script: 'git describe 2> /dev/null || exit 0').trim()
}

def getBranch() {
  def sha = getRevision()
  def branch = sh(
      returnStdout: true,
      script: "git show-ref | grep ${sha} " +
          "| grep -oP \"refs/remotes/.*/\\K.*\" | grep -v -i head || echo NOT_FOUND").trim()
  return branch == 'NOT_FOUND' ? '' : branch
}

return this