# README.md

The folder flakeyTest contains java code Pair.java, TotalFlakey.java that calculates the flakeyness (the percentage of failures for each test cases run in the past 7 and 30 days) and bash files with commands to read produced results files (junits.xml) of tests run on istio from pantheon website for branches of master and release-1.2.

The folder also contains a testCommand.sh file that holds the test commands to read from one one test suite for each branch of master and release-1.2 to reduce the amount of time it takes to read the files and run the code.

To manually run the code, one would need the dependency folder from <https://github.com/ChristinaLyu0710/istio-flakey-test/raw/master/FlakeyTest/flexible/Flakey/jars.zip> that with the libraries required in the java files.

The branches master and release-1.2 are hard coded in the .sh files to add more branches, one can edit the .sh files to simply add more gs:// links that points to the junit.xml files from other branches.
