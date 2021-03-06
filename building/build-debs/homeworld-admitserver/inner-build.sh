#!/bin/bash
set -e -u

rm -rf go
tar -xf go-bin-1.8.3.tgz go/
export GOROOT="$(pwd)/go/"
export PATH="$PATH:$GOROOT/bin"

if [ "$(go version 2>/dev/null)" != "go version go1.8.3 linux/amd64" ]
then
	echo "go version mismatch! expected 1.8.3" 1>&2
	go version 1>&2
	exit 1
fi

GODIR="$(pwd)/gosrc/"
rm -rf "${GODIR}"
ROOT=$(pwd)
mkdir "${GODIR}"

export GOPATH="${GODIR}"

(cd "${GODIR}" && tar -xf "${ROOT}/golang-x-crypto.tar.xz" src)

go build hyadmit.go token.go sshsign.go

echo "admitserver built!"
