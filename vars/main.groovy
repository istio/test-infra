/*
Run the main on master and send notification in case of error.
*/

def call(utils, Closure body) {
  node('master') {
    try {
      body()
    } catch (Exception e) {
      currentBuild.result = 'FAILURE'
      throw e
    } finally {
      utils.sendNotification(gitUtils.NOTIFY_LIST)
    }
  }
}

return this