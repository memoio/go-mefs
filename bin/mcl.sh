#!/bin/bash

mefs
if [ $? -ne 0 ]; then
        echo "====================="
        echo "recompile mcl library"
        echo "====================="
        rm -rf /mcl/build/*
        cd /mcl/build
        cmake ..
        make -j 6
        make install
        ldconfig
fi
