# Istio develop process under Mungegithub

Mungegithub is a tool based on github offering a better way to review and approve PRs and automatically rebase and merge them without engineers' efforts. Plus a two-level code-owner system to control codebase by always having the right people to approve changes. 

Mungegithub required a small behavior change to use new system while doesn't block you to keep the old way, which although is not recommended.

Submit-queue is the core part of mungegithub. Its job is to automatically rebase verify and merge PRs after proper human approvals.
Submit-queue waits for all reuqired tests passed, 4 kinds of labels(cla, lgtm, approve, release-note) set until it puts this pr into the queue. When a mergable pr comes to the head of Submit-queue, required retest contexts will be triggered to make sure everything is well tested, after these second-round tests pass, Mungegithub will go and merge this pr for you!

## Concept
Mungegitgub uses github label system to communicate with developers and other robots.

### Label
CLA:
LGTM:
Approve:
Release-note:

### Comment


### OWNERS file


## Walk Through
### Create a PR

### Review a PR

### Ready to merge

