#!/bin/bash
grep log.Print * -Rl | grep .go$ | grep -v _test.go
RESULT=$?
if [ $RESULT != 1 ]; then
    exit 1
fi

grep context.Background * -Rl | grep .go$ | grep -v _test.go
RESULT=$?
if [ $RESULT != 1 ]; then
    exit 1
fi

RESULT1=$(grep "conn, err" *.go -R | wc | awk '{print $1}')
RESULT2=$(grep "conn.Close" *.go -R | wc| awk '{print $1}')

if [ $RESULT1 != $RESULT2 ]; then
    echo "Missing Closes!"
    exit 1
fi
