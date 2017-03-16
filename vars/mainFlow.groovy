/*
Run the main on master and send notification in case of error.
*/

def call(utils, Closure body) {
  node('master') {
    def success = true
    utils.updatePullRequest('run')
    try {
      body()
    } catch (Exception e) {
      currentBuild.result = 'FAILURE'
      success = false
      throw e
    } finally {
      utils.updatePullRequest('verify', success)
      notifyList = utils.getParam(env.NOTIFY_LIST)
      if (notifyList) {
        utils.sendNotification(notifyList)
      }
    }
  }
}

return this