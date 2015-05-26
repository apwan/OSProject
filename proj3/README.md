#Projec 3

Chen Xiaoqi, Ku Lok Sun, Wu Yijie, Zhang Hanrui

Single-Server In-Memory Key-Value Server

## Source Code

In `src`: 
-  `cmap_string_string/` implements the database, 
-  `kvlib/` contains some utility functions frequently called as well as some configuration constants
-  `main/` contains the the codes of servers.

## Build

Use `compile.sh` to generate `bin` directory containing all the programs.
Use `clean.sh` to remove `bin`.

## Run

Start: `bin/start_server -p | -b &`.
Stop: `bin/stop_server -p | -b &`.

## Testing

To run the test cases (after building).
Run `bin/test`.


The following step needs to be done to make the tester work in other group's project

  1. Ensure that our `bin/test` is run in the root directory
  2. copy `conf/test.conf` to `conf/` in other project
  3. modify the parameters in `conf/test.conf`, such as the `<ip:port>` of servers 
  4. copy the whole `test/` directory (containing all the test cases) to the root directory of other group's project

Note 1: Our test case use the `bin/start_server` and `bin/stop_server`,
make sure the correct scripts is executed when our tester run in other project.

Note2: There are other parameters in `test.conf`. Setting `with_err_msg` to be true will allow the tester to print out the output of each test case. And `concur_num` specifies the number of concurrent request. If it is too large, some operations may return false beacuse the table is locked.



## Test Cases:

- 0.test: basic insert, delete and update
- 1.test: concurrent insert, delete and update
- 2.test: restart server repeatly, to test backup function
- 3.test: omitted to show that our `bin/test` is robust under I/O exception
- 4.test: test backup function, focus on primary server
- 5.test: restart server before basic insert, delete and update.
- 6.test: restart server before concurrent insert, delete and update.
- 7.test: test backup function, focus on backup server
- 8.test: try to get some records which does not exists, sequencially and concurrently
- 9.test: restart server after sequenceial and concurrent insert, delete and update.



## Implementations

### Synchonization

There is a state machine to synchronize the primary server and back up server.
The state machine have five states: 'cold start', 'warm start', 'bootstrap', 'sync' and 'shutting down'.
The server will be in 'cold start' state, if it cannot find another server.
The server will be in 'warm start' state, if it can find another server.
When the server contains data but the another does not, it will be in 'bootstrap' state.
The server go to 'shutting down' state when it receives shutdown signal.

### Data Structure

To increase the performance, only part of the table is locked for each insert or update operation.
The table is partitioned to 32 bucket. After hashing, each key will be put in exactly one bucket.
Only one bucket need to be locked when a key-value is update.

### Tester

We also carry out some pressure test. Note that the project rubric do not allow tests that exhaust memory, so we did not put them into the Tester.
We writed a tester in golang to facilitating the testing procedure. 
The testcases are put in a `.test` file in a specific format. 
The tester reads the files and fires operations to the server parallelly or sequentially.
The tester can also execute other programs and scripts.
