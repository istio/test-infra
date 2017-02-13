/*
Sets Up environment to run go.
*/

def call(gitUtils, goImportPath, Closure body) {
  def goPath = env.WORKSPACE
  def newWorkspace = "${goPath}/src/${goImportPath}"
  sh("mkdir -p ${newWorkspace}")
  withEnv(["GOPATH=${goPath}", "PATH+GOPATH=${goPath}/bin"]) {
    dir(newWorkspace) {
      body()
    }
  }
}

return this