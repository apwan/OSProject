# This is the Makefile helping you submit the labs.  
# Just create 6.824/api.key with your API key in it, 
# and submit your lab with the following command: 
#  make tester:  same-machine test

KEY=$(shell cat api.key)
SOURCES= src/main/start_server.go src/main/stop_server.go #
OURLIBS= src/paxos src/kvpaxos src/kvlib src/stoppableHTTPlistener #
OTHERS= conf test compile.sh clean.sh Makefile README.md #

project_4.tar.gz: $(SOURCES) $(OURLIBS)
	@tar czvf $@ $(SOURCES) $(OURLIBS) $(OTHERS)

test_Paxos:
	@GOPATH=`pwd` go test paxos -v

test_kvPaxos:
	@GOPATH=`pwd` go test kvpaxos -v

.PHONY: build tester clean
build:
	@./compile.sh

tester:
	@bin/test n01&
	@bin/test n02&
	@bin/test n03&
	@bin/test -m



clean:
	@-rm project_4.tar.gz
	@./clean.sh
	
