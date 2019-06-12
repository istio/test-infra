#!/bin/bash

# The file depends on the java application, 
# jars dependency folder and java files from flakeyTest folder.
# The commands in the file runs the java files from flakeyTest 
# from prow daily to calculate the percentage of flakey-ness of 
# test cases.

cd flakeyTest

/usr/local/bin/java/bin/javac -cp ".:/usr/local/bin/flakey_jars/jars/*" Pair.java TotalFlakey.java

/usr/local/bin/java/bin/java -cp ".:/usr/local/bin/flakey_jars/jars/*" TotalFlakey