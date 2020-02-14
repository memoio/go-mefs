#!/bin/bash

mefs
if [ $? -ne 0 ]; then
        echo "====================="
        echo "recompile mcl library"
        echo "====================="
        cd /go/docker-mefs-env
        make
fi
