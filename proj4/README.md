Project 4: Paxos-based Key-Value Service
====

# Design

Three servers in library `kvpaxos` runs with underlying `paxos` library.

## Paxos

The `paxos` library will create paxos instances, which will try to decide an operation for each slot. The decision will contain the data (key/value)  and operation type (Put, Update, etc.) and is consistent for majority.

This library has been fully tested using the original paxos test.

## KVPaxos

The `kvpaxos` library will use its underlying `paxos` instance to achieve consistent database service. Upon each request (including both KV service and KVMAN service), one decision slot is obtained and the operation logged into paxos history. All data queries will be recovered from the database log.

In order to improve performance, we used a random-backoff scheme (similar to that in CSMA/CD) and optimized the inter-arrival time of paxos decision queue, such that each decision can be quickly made even if the system is under high load pressure.

The client may optionally provide an operation ID, and the server will not repeat an operation with the same ID appeared before. This will ensure database consistency in the case of server temporary partition and client/server unreliable communication.

Each server will periodically create snapshots of database log (and let paxos forget old decisions), to help reduce memory comsumption and improve performance (of scanning the log). The threshold to trigger snapshot update can be modified in the configuration. Note that in order to pass the original memory consumption test, the threshold muss be less than 50 (since there's only about 50 operations in the test).

This library has been fully tested using the original kvpaxos client/server test, before migrating to the new RPC platform.

## RPC Interface

To comply with the multi-machine scenario, we changed the communication between different paxos instances to TCP-based (from Unix-socket based). This does not affect the normal working of paxos, since the RPC and the network transportation is fully layered; however, now we cannot control the partition in test cases.

## HTTP Interface

Each kvpaxos instance will listen to a HTTP port, to provide the following service:

### Data service
#### Put `/kv/insert` or `/kv/put`
Insert new key into the database; will succeed only if it's a new key. Require the key to be nonempty.

The parameter is provided in `key` and `value` field.

#### Update `/kv/update`
Update key in the database; will succeed only if it's an existing key.
The old value will be returned.

The parameter is provided in `key` and `value` field.

#### Delete `/kv/delete` 
Delete key in the database; will succeed only if it's an existing key.
The old value will be returned.

The parameter is provided in `key`  field.

#### Get `/kv/get` 
Look up a key in the database; will succeed only if it's an existing key.
The value will be returned.

The parameter is provided in `key`  field.

Note: Each HTTP request is treated as independent requests, since the HTTP protocol is stateless; if consistency in unreliable network is desired, the client should provide a unique increasing operation ID in `opid` field, and the server will not repeat multiple requests with the same ID or an older ID.

### Management service
#### CountKey `/kvman/countkey` 
Returns the number of distinct, existing keys in the database. 

This operation will succeed only if the server can obtain an agreement (i.e. not partitioned into minority) such that the data is up to date.

#### Dump `/kvman/dump` 
Returns a list of existing key-value pairs in the database. 

This operation will succeed only if the server can obtain an agreement (i.e. not partitioned into minority) such that the data is up to date.

#### Shutdown `/kvman/shutdown` 
Kills the server and release the listening ports.

Note: this operation may require at most 10 milliseconds timeout, before the ports are released and a new instance can be started. Listening to the port immediately after shutdown may cause an error. This issue is handled in our server initialization process, by waiting 10ms before starting any port; however it may affect a subsequent group's project if projects of many groups is tested in batch automatically.


## Tester

This project uses a similar tester in previous project, which initializes HTTP requests to different servers and send different requests according to test case files; the result of each request is automatically deduced from previous request sequence and compared with actual result. The test will pass only if all result matched.

Normally, the servers are started before each case and shut after the test finishes. Optionally, a test case can choose to shut down one or more servers during the testing process. In this case the correct results are also automatically deduced using the majority consensus requirement. Due to aforementioned difficulty, the test case does not implement partition.

To help testing on multiple servers, we implementd auxiliary test helpers, who runs on different machines and help main tester to start and stop kvpaxos instances.

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
