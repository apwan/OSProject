Project 4: Paxos-based Key-Value Service
====

# Design

Including `paxos`, `kvpaxos`, `kvlib`.

# Build and Run

Use `compile.sh`, `bin/start_server` and `bin/stop_server`.


#Test
Using the same config files (`settings.conf` and `test.conf`), in each server, run the corresponding remote tester:
```
bin/test 1 &

bin/test 2 &

bin/test 3 &
```

Then in any machine, run the main tester:
```
bin/test -m
```
