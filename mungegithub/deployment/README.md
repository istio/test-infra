# Istio develop process under Mungegithub

Mungegithub is a tool based on github offering a better way to review and approve PRs and automatically rebase and merge them without engineers' efforts. Plus a two-level code-owner system to control codebase by always having the right people to approve changes. 

Mungegithub required a small behavior change to use new system while doesn't block you to keep the old way, which although is not recommended.

Submit-queue is the core part of mungegithub. Its job is to automatically rebase verify and merge PRs after proper human approvals.
Submit-queue waits for all reuqired tests passed, 4 kinds of labels(cla, lgtm, approve, release-note) set until it puts this pr into the queue. When a mergable pr comes to the head of Submit-queue, required retest contexts will be triggered to make sure everything is well tested, after these second-round tests pass, Mungegithub will go and merge this pr for you!

## Concept
Mungegitgub uses github label and comment system to communicate with developers and other robots.

### Comment
Powered by prow plugins and mungegithub, people can add labels to PRs by directly write comment. 

  >  Note: If not necessary, do not add or remove labels from github UI which will bypass Mungegithub access control system.

### Label
Mungegithub waits for 4 labels:
* **CLA** "cla-yes" label or "cla-no" label is set automatically. There is a hacker way "cla human-approved" to bypass cla check if it's necessary.
* **LGTM** "lgtm" is the first level approve, it means "look good to me, but someone else may need to take a look and make final decision". "lgtm" is more like "review approve" in github-way. Everyone assigned to this pr can say valid "/lgtm", people in github organization could self-assign. With prow deployed, people should add "lgtm" label by comment "/lgtm". 

  >  Note: Any code change after "/lgtm" will automatically remote "lgtm" label
  
* **Approve** "approved" is the second level approve, it's more like clicking the "merge" buttom in github-way. So when you are a repo admin or the owner of this part of code, when you actually want to get this pr into master, instead of clicking the merge buttom, simply comment "/approve" (We are getting rid of "/approve no-issue"), Mungegithub will add "approved" label on the pr and when other requirements are satisfied, this pr is going to the submit-queue.

  >  Note: Do not comment "/approve" or add "approved" label unless you are 100% sure you want this change. 
  
* **Release-note** Release note enforcement is another feather we are seeking from Mungegithub. With template, when prs are create, people should add release note (can be "None") in the pr description. Depended on the release message left, Mungegithub will add "release-note", "release-note-none". If you leave it empty, "release-note-needed" will be added and is going to block merging.

* **do-not-merge** Even with all required parts, you can always have more time by add "do-not-merge" label. 

### OWNERS file
OWNERS file is the way to organize code owners and write priviliage. There are two parts. Reviews are the people who are suggested to review the pr and approvers are the people who can actually say "/approve" to add the "approve" label. Each path can have its OWNERS file, and this file affects this directory and all subdirectories. More details can be found: [Reviewer and approver](https://github.com/kubernetes/test-infra/tree/master/mungegithub/mungers/approvers)

## Walk Through
### Create a PR

### Review a PR

### Ready to merge


# Related links:

