# Istio develop process under Mungegithub

  >  Note: This doc is for normal processes of Mungegithub in istio organization. Exceptions may apply to experimental environments.

  Currently deployed: istio, mixer, auth, test-infra, pilot
  
  Limited functional (only merge auto-dependency-update PRs): mixerclient, proxy

Mungegithub is a tool based on GitHub that offers a better way to review and approve PRs and automatically rebases and merges them without engineers' effort. Plus a two-level approval system controls codebase better by always having the right people to approve changes. 

Mungegithub requires a behavior change to use new system while doesn't block you to keep the old way, which although is not recommended.

The core part of mungegithub, submit-queue automatically rebases, verities and merges PRs once they are approved.
It waits for all reuqired tests passed, 4 kinds of labels(cla, lgtm, approve, release-note) before putting this PR into the queue. When a mergable PR comes to the head of Submit-queue, required retest contexts will be triggered to make sure everything is well tested. After these second-round tests pass, Mungegithub will go and merge this PR for you!

## Concept
Mungegitgub uses GitHub label and comment system to communicate with developers and other robots.

### Comment
Powered by prow plugins and mungegithub, people can add labels to PRs by commentting on PRs. 

  >  Note: If not necessary, do not add or remove labels from GitHub UI which bypasses Mungegithub access control system.

### Labels
Mungegithub waits for 4 labels:

* **CLA** "cla-yes" label or "cla-no" label is set automatically. There is a hacker way "cla human-approved" to bypass cla check if it's necessary.
* **LGTM** "lgtm" is the first level approve, it means "look good to me, but someone else may need to take a look and make final decision". "lgtm" is more like "review: approve" in github-way. Everyone assigned to this PR can say valid "/lgtm", people in GitHub organization can also do it. With prow deployed, people should add "lgtm" label by comment "/lgtm". 

  >  Note: Any code changes after "/lgtm" will automatically remove "lgtm" label
  
* **Approve** "approved" is the second level approve, it's like "merge" button in GitHub-way. 
As the repo admin or the directory owner, one may approve and merge this PR into master by simply commenting "/approve", after which Mungegithub will add an "approved" label on the PR. Permitted all other merging requirements are satisfied, this PR is put on the submit-queue.

  >  Note: Do not comment "/approve" or add "approved" label unless you are 100% sure you want this change, because after you say that, the PR will be merge any minutes.
    
* **Release-note** Release note enforcement is another feature we are seeking from Mungegithub. With template, when PRs are created, people should add release note (can be the word "none") in the PR description. Depended on the release message left, Mungegithub will add label "release-note-none"(if you write string `none`) or "release-note-action-required"(if you write string `action required`) or "release-note"(Any other message). If you leave it empty, **"do-not-merge/release-note-label-needed"** will be added and will block merging. At release point, a tool will gather release-note messages from PRs with "release-note" label.

  > __New change after [PR#531](https://github.com/istio/test-infra/pull/531):__ _Mungegithub is no longer handling "release-note" labels, instead, Prow takes care of that. And good news is after you added "release-note" in PR description, Prow will automatically remove "do-not-merge/release-note-label-needed" and will unblock merge processes._

* **do-not-merge** Merge process will be blocked in any stage due to the existence of any kinds of "do-not-merge" labels.

  * _"do-not-merge"_: normal label to stop merge, can be added for general reason by people with write access.
  * _"do-not-merge/hold"_: It's a easy way for everyone to hold merge, if you simply comment **"/hold"** or **"/hold cancel"**, robot will add or remove "do-not-merge/hold" and it will block/unblock merge process.
  * _"do-not-merge/release-note-label-needed"_: This label is added by Prow due to the missing release-note part in PR description. 
  As long as release-note message is filled, this label will be removed automatically.
  

### OWNERS file
OWNERS file is the way to organize code owners and write priviliage. There are two parts. Reviews are the people who are suggested to review the PR and approvers are the people who can actually say "/approve" to add the "approve" label. Each path can have its OWNERS file, and this file affects this directory and all subdirectories. More details can be found: [Reviewer and approver](https://github.com/kubernetes/test-infra/tree/master/mungegithub/mungers/approvers)

## Auto-merge a PR with Mungegithub

### Stage One

#### 1. Required CI Required CI status must be green. Due to the lack of GitHub api, Mungegithub gets required CI from configration.
[labels.png]

#### 2. Four kinds of Labels "cla: yes", "lgtm", "approve", "release-note"/"release-note-none"
![ci-status](https://github.com/istio/test-infra/blob/master/mungegithub/images/ci-status.png)

* **cla** 

  - Google-bot will add cla label, if you get a "cla: no" label, follow the instruction offered by googlebot to signup cla. 

![cla-label](https://github.com/istio/test-infra/blob/master/mungegithub/images/cla-label.png)

* **lgtm** 

  - Anyone in the organization can say `/lgtm` to add "lgtm" label.  
  - No one can say `/lgtm` to his/her own PR.  
  - When a valid approver says `/lgtm`, it implies `/approve` as well.  
  - You can cancel "lgtm" by commenting `/lgtm cancel`, if you `/lgtm` before.

![lgtm](https://github.com/istio/test-infra/blob/master/mungegithub/images/lgtm.png)

* **approved**

  - When a person's name is in all OWNERS file(s) which is/are able to cover all changed files, he/she is a valid approver.
A valid approver is able to say `/approve` and mungegithub will add "approved" label.  
  - We already got rid of `/approve no-issue`, every PR can be approved by just saying `/approve`  
  - A valid approver can say `/approve` to his/her own PR  
  - You can cancel "approved" by commenting "/approve cancel", if you `/approve` before.

![approve](https://github.com/istio/test-infra/blob/master/mungegithub/images/approve.png)

* **release-note**
  - Use [PR template](https://github.com/istio/istio/blob/master/.github/PULL_REQUEST_TEMPLATE.md) to add release-note, make sure to add "none" if release-note isn't necessary for this PR.
  
  ![release-note template](https://github.com/istio/test-infra/blob/master/mungegithub/images/release-note-template.png)
  
  ~~- You can also leave a comment `/release-note-none` to set it NONE~~
  
  > ~~When all requirements are satisfied except "release-note", mungegithub will add "do-not-merge" label to force you clear release-note. Please note, when you fix release-note, you have to manually remove that label after you make sure everything looks good.~~
  
  - If you put nothing here, Prow is going to complain about it by adding "do-not-merge/release-note-label-needed" 
  ![do-not-merge:release-note-label-needed](https://github.com/istio/test-infra/blob/master/mungegithub/images/do-not-merge:release-note-label-needed.png)
  
  - Additional if you put **"action required"**, robot will add "release-note-action-required". This is a **valid merge label**. But we need to make sure we come back and add actual release-note for PRs have this label later on (before next release).
  ![release-note-action-required](https://github.com/istio/test-infra/blob/master/mungegithub/images/release-note-action-required.png)
  
* **do-not-merge**
  - If you want to block merge, comment "/hold", and when you think it's ready, comment "/hold cancel"
  ![do-not-merge:hold](https://github.com/istio/test-infra/blob/master/mungegithub/images/do-not-merge:hold.png)
  
### Stage Two
  When a PR satisfies all requirements in stage one, it will be picked up by submit-queue. Submit-queue is going to sequentially retest each PR by asking prow to rerun all required tests. If, during reruning tests, any required labels are removed or any kind of "do-not-merge" label is added, (e.g. someone comments "/hold"), the tests won't abort but the merge will be blocked for sure.
  ![retest](https://github.com/istio/test-infra/blob/master/mungegithub/images/retest.png)
  - Idealy, you don't need to update branch (also called rebase/merge from master) before running tests because prow will merge your branch into master locally and run tests against this code. You will not see any changes happens on your branch but you can find some clue when you go to prow log.
  ![auto-rebase-log](https://github.com/istio/test-infra/blob/master/mungegithub/images/auto-rebase-log.png)
  - Use `/test all` or `/retest`, you can manually verify and retest everything.
  
 ### Stage Three
   If a PR passes all required tests in rerun, then this PR is good to go and will be auto merged by submit-queue, wohoo!
   ![merge](https://github.com/istio/test-infra/blob/master/mungegithub/images/merge.png)
 
 
 ## FAQ
 
* Why my PR hasn't been retest
   - Four kinds of labels
   - No "do-not-merge" label
   - All required checks pass
   - Maybe another PR is being retested, yours are waiting in the queue

* Why does my `/lgtm`, adds both "lgtm" and "approved" labels
   - If you are a valid approvor, your `/lgtm` implies `/approve`
