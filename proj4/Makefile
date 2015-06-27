# This is the Makefile helping you submit the labs.  
# Just create 6.824/api.key with your API key in it, 
# and submit your lab with the following command: 
#  make run:  same-machine test

KEY=$(shell cat api.key)
SOURCES= src/main/start_server.go src/main/stop_server.go #
OURLIBS= src/paxos src/kvpaxos src/kvlib
OTHERS= conf test compile.sh clean.sh Makefile README.md #

project_4.tar.gz: $(SOURCES) $(OURLIBS)
	tar czvf $@ $(SOURCES) $(OURLIBS) $(OTHERS)

build:
	./compile.sh

run:
	@bin/test 1&
	@bin/test 2&
	@bin/test 3&
	@bin/test -m

test_Paxos:
	@GOPATH=`pwd` go test paxos -v

test_kvPaxos:
	@GOPATH=`pwd` go test kvpaxos -test.run TestPartitioin -v


clean:
	-rm project_4.tar.gz
	
