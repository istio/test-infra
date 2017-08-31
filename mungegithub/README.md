# Istio develop process under Mungegithub

  >  Note: This doc is for normal processes of Mungegithub in istio organization. Exceptions may apply to experimental environments.

  Current deployed: istio, mixer, auth, test-infra 
  Current paused: pilot 

Mungegithub is a tool based on github offering a better way to review and approve PRs and automatically rebase and merge them without engineers' efforts. Plus a two-level approval system controls codebase better by always having the right people to approve changes. 

Mungegithub requires a behavior change to use new system while doesn't block you to keep the old way, which although is not recommended.

Submit-queue is the core part of mungegithub. Its job is to automatically rebase, verify and merge PRs after proper human approvals.
Submit-queue waits for all reuqired tests passed, 4 kinds of labels(cla, lgtm, approve, release-note) set until it puts this pr into the queue. When a mergable pr comes to the head of Submit-queue, required retest contexts will be triggered to make sure everything is well tested, after these second-round tests pass, Mungegithub will go and merge this pr for you!

## Concept
Mungegitgub uses github label and comment system to communicate with developers and other robots.

### Comment
Powered by prow plugins and mungegithub, people can add labels to PRs by directly write comment. 

  >  Note: If not necessary, do not add or remove labels from github UI which bypasses Mungegithub access control system.

### Label
Mungegithub waits for 4 labels:
* **CLA** "cla-yes" label or "cla-no" label is set automatically. There is a hacker way "cla human-approved" to bypass cla check if it's necessary.
* **LGTM** "lgtm" is the first level approve, it means "look good to me, but someone else may need to take a look and make final decision". "lgtm" is more like "review: approve" in github-way. Everyone assigned to this pr can say valid "/lgtm", people in github organization can also do it. With prow deployed, people should add "lgtm" label by comment "/lgtm". 

  >  Note: Any code changes after "/lgtm" will automatically remove "lgtm" label
  
* **Approve** "approved" is the second level approve, it's more like clicking the "merge" buttom in github-way. So when you are a repo admin or the owner of this part of code, when you actually want to get this pr into master, instead of clicking the merge buttom, simply comment "/approve", Mungegithub will add "approved" label on the pr and when other requirements are satisfied, this pr is going to the submit-queue.

  >  Note: Do not comment "/approve" or add "approved" label unless you are 100% sure you want this change, because after you say that, the pr will be merge any minutes.
    
* **Release-note** Release note enforcement is another feather we are seeking from Mungegithub. With template, when prs are create, people should add release note (can be "None") in the pr description. Depended on the release message left, Mungegithub will add "release-note", "release-note-none". If you leave it empty, "release-note-needed" will be added and is going to block merging.

* **do-not-merge** Even with all required parts, you can always have more time by add "do-not-merge" label. 

### OWNERS file
OWNERS file is the way to organize code owners and write priviliage. There are two parts. Reviews are the people who are suggested to review the pr and approvers are the people who can actually say "/approve" to add the "approve" label. Each path can have its OWNERS file, and this file affects this directory and all subdirectories. More details can be found: [Reviewer and approver](https://github.com/kubernetes/test-infra/tree/master/mungegithub/mungers/approvers)

## Auto-merge a PR with Mungegithub

### Stage One

#### 1. Required CI Required CI status must be green. Due to the lack of github api, Mungegithub gets required CI from configration.
[labels.png]

#### 2. Four kinds of Labels "cla: yes", "lgtm", "approve", "release-ntoe"/"release-note-none"
![alt tag](https://github.com/istio/test-infra/blob/mungegithub_doc/mungegithub/images/ci-status.png)

* **cla** 

  - Google-bot will add cla label, if you get a "cla: no" label, follow the instruction offered by googlebot to signup cla. 

![alt tag](https://github.com/istio/test-infra/blob/mungegithub_doc/mungegithub/images/cla-label.png)

* **lgtm** 

  - Anyone in the organization can say `/lgtm` to add "lgtm" label.  
  - No one can say `/lgtm` to his/her own PR.  
  - When a valid approver say `/lgtm`, it implies `/approve` as well.  
  - You can cancel "lgtm" by commenting `/lgtm cancel`, if you `/lgtm` before.

![alt tag](https://github.com/istio/test-infra/blob/mungegithub_doc/mungegithub/images/lgtm.png)

* **approved**

  - When a person's name is in all OWNERS file(s) which is/are able to cover all changed files, he/she is a valid approver.
A valid approver is able to say `/approve` and mungegithub will add "approved" label.  
  - We already got ride of `/approve no-issue`, every PR can be approved by just saying `/approve`  
  - A valid approver can say `/approve` to his/her own PR  
  - You can cancel "approved" by commenting "/approve cancel", if you `/approve` before.

![alt tag](https://github.com/istio/test-infra/blob/mungegithub_doc/mungegithub/images/approve.png)

* **release-note**
  - Use [PR template](https://github.com/istio/istio/blob/master/.github/PULL_REQUEST_TEMPLATE.md) to add release-note, make sure add "none" if that's not necessary for this PR.
  
  ![alt tag](https://github.com/istio/test-infra/blob/mungegithub_doc/mungegithub/images/release-note-template.png)
  
  - You can also leave a comment `/release-note-none` to set it NONE
  > When all requirements are satisfied except "release-note", mungegithub will add "do-not-merge" label to force you clear release-note. Please note, when you fix release-note, you have to manually remove that label after you make sure everything looks good.
  ![alt tag](https://github.com/istio/test-infra/blob/mungegithub_doc/mungegithub/images/do-not-merge-release-note.png)
  
* **do-not-merge**
  - "do-not-merge" label will block merge anytime. Add it if you want to temporarily block merge.
  
### Stage Two
  When a PR satisfies all requirements in stage one, it will be picked up by submit-queue. Submit-queue is going to sequentially retest each pr by asking prow to rerun all required tests.
  ![alt tag](https://github.com/istio/test-infra/blob/mungegithub_doc/mungegithub/images/retest.png)
  - Idealy, you don't need to update branch (also called rebase/merge from master) before run tests because prow will merge your branch into master locally and run tests against this code. You will not see any changes happens on your branch but you can find some clue when you go to prow log.
  ![alt tag](https://github.com/istio/test-infra/blob/mungegithub_doc/mungegithub/images/auto-rebase-log.png)
  - Use `/test all` or `/retest`, you can manually verify and retest everything.
  
 ### Stage Three
   If a PR passes all required tests in rerun, then this PR is good to go and will be auto merged by submit-queue, wohoo!
   ![alt tag](https://github.com/istio/test-infra/blob/mungegithub_doc/mungegithub/images/merge.png)
 
 
 ## FAQ
 
   * Why my pr hasn't been retest
   - Four kinds of labels
   - No "do-not-merge" label
   - All required checks pass
   - Maybe another PR is being retested, yours are waiting in the queue

   * Why I said `/lgtm`, it also adds "approved" label
   - If you are a valid approvor, your `/lgtm` implies `/approve`
