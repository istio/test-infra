/*
Creates a node with the right label and checkout the source code.
*/

import org.istio.testutils.Utilities


def call(gitUtils, goImportPath, Closure body) {
  utils = new Utilities()
  def nodeLabel = utils.getParam('SLAVE_LABEL', gitUtils.DEFAULT_SLAVE_LABEL)
  def buildNodeLabel = "${nodeLabel}-build"
  node(buildNodeLabel) {
    env.GOPATH = env.WORKSPACE
    env.PATH = "${env.GOPATH}/bin:${env.PATH}"
    def newWorkspace = "${env.GOPATH}/src/${goImportPath}"
    sh("mkdir -p ${newWorkspace}")
    dir(newWorkspace) {
      gitUtils.checkoutSourceCode()
      body()
    }
  }
}
