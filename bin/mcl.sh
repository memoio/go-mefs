#!/bin/bash

mefs
if [ $? -ne 0 ]; then
        echo "====================="
        echo "recompile mcl library"
        echo "====================="
        rm -rf /go/src/mcl/mcl/build/*
        cd /go/src/mcl/mcl/build
        cmake ..
        make -j 6
        make install
        ldconfig
fi
