#!/usr/bin/env bash
set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

cd ${SCRIPT_DIR}
rm -rf go.mod go.sum
go mod init tracer
go mod edit -replace github.com/caesurus/riptracer=../../../riptracer
go mod tidy
go clean 
echo "Building go code"
GOARCH=386 GOOS=linux go clean
GOARCH=386 GOOS=linux go build

echo "Run the application via start"
./tracer start -c ./vuln

# should see: pid:xxxxxx -> doNothing called with arg: -721750240
