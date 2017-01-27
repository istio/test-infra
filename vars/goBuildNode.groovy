/*
Creates a node with the right label and checkout the source code.
*/

def call(gitUtils, goImportPath, Closure body) {
  def nodeLabel = params.get('SLAVE_LABEL')
  if (nodeLabel == null) {
    nodeLabel = gitUtils.DEFAULT_SLAVE_LABEL
  }
  def buildNodeLabel = "${nodeLabel}-build"
  node(buildNodeLabel) {
    def goPath = env.WORKSPACE
    def path = "${goPath}/bin:${env.PATH}"
    def newWorkspace = "${goPath}/src/${goImportPath}"
    sh("mkdir -p ${newWorkspace}")
    withEnv(["GOPATH=${goPath}", "PATH=${path}"]) {
      dir(newWorkspace) {
        gitUtils.checkoutSourceCode()
        body()
      }
    }
  }
}

return this