#!/usr/bin/env bash
# should be executed under root directory
ROOT_DIR=`pwd`;
prog_src=(
src/main/start_server
src/main/stop_server
test/test
)
echo "start compiling ...";
mkdir bin; cd bin;
for j in ${prog_src[*]}; do
	GOPATH=${ROOT_DIR} go build ${ROOT_DIR}/$j.go;
done
echo "compiled.";
