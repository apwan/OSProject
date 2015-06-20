#!/bin/bash
killall server
#start server
bin/start_server -p > /dev/null &
bin/start_server -b > /dev/null &
sleep 1

PORT=8088 # port used by primrary server
BPORT=8089 # port used by backup
echo $PORT
echo $BPORT

# Insert 10 pairs
echo "Insert 10 pairs"
curl localhost:$PORT/kv/insert -d "key=r3gdcMfAgp&value=mwhxFGugTs"
curl localhost:$PORT/kv/insert -d "key=ZxqU7FLnno&value=JpGkTgc4gm"
curl localhost:$PORT/kv/insert -d "key=sYuZ3Aavss&value=9INoiYXgAU"
curl localhost:$PORT/kv/insert -d "key=mczOmXJ0zH&value=3frX0VHTxs"
curl localhost:$PORT/kv/insert -d "key=bl4aMK3SuF&value=hLytNeOZYa"
curl localhost:$PORT/kv/insert -d "key=7SJiw0huh1&value=hF7epB6wJl"
curl localhost:$PORT/kv/insert -d "key=eWPyEVVnGJ&value=zKOOh5V6zv"
curl localhost:$PORT/kv/insert -d "key=hPaJW9fOit&value=xUuw62iU4s"
curl localhost:$PORT/kv/insert -d "key=726XVoqqBr&value=ZlFOk7VbVe"
curl localhost:$PORT/kv/insert -d "key=TvXffVy6Ys&value=8vdcaYOY1j"

# Read them back
echo "Read 10 pair"
curl localhost:$PORT/kv/get?key=r3gdcMfAgp
echo
curl localhost:$PORT/kv/get?key=ZxqU7FLnno
echo
curl localhost:$PORT/kv/get?key=sYuZ3Aavss
echo
curl localhost:$PORT/kv/get?key=mczOmXJ0zH
echo
curl localhost:$PORT/kv/get?key=bl4aMK3SuF
echo
curl localhost:$PORT/kv/get?key=7SJiw0huh1
echo
curl localhost:$PORT/kv/get?key=eWPyEVVnGJ
echo
curl localhost:$PORT/kv/get?key=hPaJW9fOit
echo
curl localhost:$PORT/kv/get?key=726XVoqqBr
echo
curl localhost:$PORT/kv/get?key=TvXffVy6Ys
echo

# Restart backup on successful restart
echo "restart backup"
# kill -9 $(lsof -i:$BPORT -sTCP:LISTEN -t)
curl localhost:$BPORT/kvman/shutdown
echo
sleep 1
bin/start_server -b > /dev/null &
sleep 1


# Delete 2 pair – without error returne 5%
echo "delete 2 pairs"
curl localhost:$PORT/kv/delete -d "key=r3gdcMfAgp"
echo
curl localhost:$PORT/kv/delete -d "key=ZxqU7FLnno"
echo

# Update 2 pair, read back the results 5%
echo "update 2 pairs"
curl localhost:$PORT/kv/update -d "key=sYuZ3Aavss&value=wFZoAUjxsX"
echo
curl localhost:$PORT/kv/update -d "key=mczOmXJ0zH&value=zjjArvvUuk"
echo

# Restart primary, on successful restart – 5%
kill -9 $(lsof -i:$PORT -sTCP:LISTEN -t)
# curl localhost:$PORT/kvman/shutdown
echo
sleep 1
bin/start_server -p > /dev/null &
sleep 1


# Dump all key-values, and check with desired results – 35%
curl localhost:$PORT/kvman/dump
curl localhost:$PORT/kvman/shutdown
curl localhost:$BPORT/kvman/shutdown
