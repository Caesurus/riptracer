#!/usr/bin/env bash
set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

cd ${SCRIPT_DIR}/c_src
echo "Compiling C application"
make clean && make 

OFFSET=$(objdump -D test_threads|grep \<fflush@plt\>|grep call|cut -d ':' -f1|xargs)
OFFSET=$(objdump -D test_threads|grep \<foo\>:|cut -d ' ' -f1|xargs)
FOO2OFFSET=$(objdump -D test_threads|grep \<foo2\>:|cut -d ' ' -f1|xargs)

cd ${SCRIPT_DIR}
rm -rf go.mod go.sum
go mod init tracer
go mod edit -replace github.com/caesurus/riptracer=../../../riptracer
go mod tidy
go clean 
echo "Building go code"
go build

if [ "$EUID" -eq 0 ]; then
    echo "Run the application via attach"
    ./c_src/test_threads &
    PID=$!
    ./tracer attach --breakpoint "${OFFSET}" -p ${PID}
    echo "Done"
fi

echo "Run the application via start"
./tracer start --breakpoint "${OFFSET}" --hwbreakpoint "${FOO2OFFSET}" -c ./c_src/test_threads
