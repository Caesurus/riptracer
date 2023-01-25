#!/usr/bin/env bash
set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

cd ${SCRIPT_DIR}
echo "Downloading crackme CTF challenge"
wget https://github.com/Caesurus/usercorn_examples/raw/master/2018_angstrom_product_key/activate -O activate
chmod +x activate

echo "Initializing go module"
rm -rf go.mod go.sum
go mod init activate_crack
go mod edit -replace github.com/caesurus/rip_tracer=../../../rip_tracer
go mod tidy
go clean 
echo "Building go code"
go build
echo "Build successful, running activate crack against"
cat user_input.txt |./activate_crack start -c ./activate

# should see "Valid Serial Key = 3914-6104-4611-1711-1243-4699" as the output

