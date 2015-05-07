#!/usr/bin/env bash
echo "start compiling ...";
mkdir build; cd build;
go build ../src/server.go;
go build ../test/client.go;
echo "compiled.";