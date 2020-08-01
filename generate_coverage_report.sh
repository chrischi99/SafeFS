#!/bin/bash

if [ $# != 1 ]; then
echo "USAGE: need a name"
exit 1;
fi

go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o ./tmp.html
cp ./tmp.html $1
