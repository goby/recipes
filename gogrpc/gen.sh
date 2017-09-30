#!/bin/bash

LIST=`find . -name '*.proto' | xargs dirname | sort`

SRC_ROOT=$PWD

for i in $LIST; do
    echo "Processing $i/"
    cd $PWD/$i
    protoc --go_out=plugins=grpc:. *.proto
done
