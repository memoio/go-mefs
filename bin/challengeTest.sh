#!/bin/bash

sh ./mcl.sh

echo -e "mefs challenge test"

echo -e "\nstep 1,mefs init\n"
mefs init --netKey=$2

echo -e "\nstep 2,run mefs daemon\n"
mefs daemon --netKey=$2 >> ~/daemon.stdout 2>&1 &
echo -e "\ndaemon is ready wait 1min to connect"
time sleep 60

echo -e "\nstep 3,run challenge test\n"
GO111MODULE=off go run $GOPATH/src/github.com/memoio/go-mefs/test/chalRepair/test.go -eth=$1
