/*
Creates a node with the right label and checkout the source code.
*/

import org.istio.testutils.Utilities

def call(gitUtils, Closure body) {
  utils = new Utilities()
  def nodeLabel = utils.getParam('SLAVE_LABEL', gitUtils.DEFAULT_SLAVE_LABEL)
  def buildNodeLabel = "${nodeLabel}-build"
  node(buildNodeLabel) {
    gitUtils.checkoutSourceCode()
    body()
  }
}