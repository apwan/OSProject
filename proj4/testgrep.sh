#!/usr/bin/env bash
# grep test output to improve readability
PAXOS_TESTS=(
TestBasic
TestDeaf
TestForget
TestManyForget
TestForgetMem
TestRPCCount
TestMany
TestOld
TestManyUnreliable
TestPartition
TestLots
)

KVPAXOS_TESTS=(
TestBasic
TestDone
TestPartition
TestUnreliable
TestHole
TestManyPartition
)

TESTS_TO_RUN=(
TestBasic
TestDeaf
)

for ii in ${KVPAXOS_TESTS[*]}; do
	echo "Running ${ii}"
	GOPATH=`pwd` go test paxos -test.run ${ii} -v > tmp.out 2>&1
	if ! (grep -q Pass tmp.out) 
	then
		echo "${ii} Failed" >> $1
	else
		echo "${ii} Passed" >> $1
	fi
done

rm *.out



