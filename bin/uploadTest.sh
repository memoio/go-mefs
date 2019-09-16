#!/bin/bash

echo -e "mefs upload test"

echo -e "\nstep 1,mefs init\n"
mefs init

echo -e "\nstep 2,run mefs daemon\n"
mefs daemon --netKey=$2 >> ~/daemon.stdout 2>&1 &
echo -e "\ndaemon is ready wait 1min to connect"
time sleep 60

echo -e "\nstep 3,run challenge test\n"
cd $GOPATH/src/github.com/memoio/go-mefs/test/unit/uploadTest
go run uploadTest.go -count=$3 -eth=$1
