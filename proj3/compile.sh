#!/usr/bin/env bash
# should be executed under root directory
ROOT_DIR=`pwd`;
prog_src=(
src/main/server
test/client
OSTester/regularTester
)
echo "start compiling ...";
mkdir build; cd build;
for j in ${prog_src[*]}; do
	GOPATH=${ROOT_DIR} go build ${ROOT_DIR}/$j.go;
done
echo "compiled.";
