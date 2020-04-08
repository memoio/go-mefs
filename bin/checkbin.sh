#!/bin/bash

mefs-user
if [ $? -ne 0 ]; then
        echo "====================="
        echo "recompile mcl library"
        echo "====================="
        cd /go/docker-mefs-env
        make install-mcl-ull
fi
