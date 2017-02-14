/*
Run the main on master and send notification in case of error.
*/

def call(utils, Closure body) {
  node {
    try {
      body()
    } catch (Exception e) {
      currentBuild.result = 'FAILURE'
      throw e
    } finally {
      notifyList = utils.getParam(env.NOTIFY_LIST)
      if (notifyList) {
        utils.sendNotification(notifyList)
      }
    }
  }
}

return this