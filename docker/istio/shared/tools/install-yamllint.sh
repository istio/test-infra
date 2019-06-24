#!/bin/bash

set -eux


apt-get -qqy install python3-pip  
pip3 install --upgrade pip
pip install --user yamllint 
export PATH=${PATH}:/root/.local/bin