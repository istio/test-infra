# Job Config Management

Presubmit and Postsubmit configs are separated per org/repo/branch to ease
config management.

All periodic jobs are defined in the all-periodics.yaml.

Note that all the files in this directory must have a unique file name. A test
will make sure this is the case. We only allow duplicate job if they run on
different branches.

In order to create a new branch, copy the master.yaml file to the branch name of
your choice, and update the jobs branch spec in this file to point to your new
branch.

