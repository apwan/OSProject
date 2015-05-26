#!/usr/bin/env bash
# should be executed under root directory
ROOT_DIR=`pwd`;
prog_src=(
src/main/server
src/main/stop_server
test/test
test/pressure_naive
)
echo "start compiling ...";
mkdir build; cd build;
for j in ${prog_src[*]}; do
	GOPATH=${ROOT_DIR} go build ${ROOT_DIR}/$j.go;
done
echo "compiled.";
