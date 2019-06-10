
#!/bin/bash

set -eux

curl -L -C - -b oraclelicense=accept-securebackup-cookie -O http://download.oracle.com/otn-pub/java/jdk/8u131-b11/d54c1d3a095b4ff2b6607d096fa80163/jdk-8u131-linux-x64.tar.gz
tar zxvf jdk-8u131-linux-x64.tar.gz
mkdir /usr/local/bin/java
cp -frp jdk1.8.0_131/* /usr/local/bin/java

export PATH=${PATH}:/usr/local/bin/java/bin