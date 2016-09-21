#!/bin/bash
# Fail on any error.
set -e
# Display commands being run.
set -x

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
MIXO="/tmp/go/src/somnacin-internal/mixologist"

echo 'Installing dependencies'
apt-get install google-golang golang-go.tools make

rm -rf /tmp/go && mkdir -p /tmp/go/{bin,src/somnacin-internal,pkg}

echo 'Installing glide'
pushd /tmp/go/bin
wget https://github.com/Masterminds/glide/releases/download/v0.11.1/glide-v0.11.1-linux-amd64.tar.gz -O glide.tar.gz
tar zxvf glide.tar.gz linux-amd64/glide --strip 1
rm -rf glide.tar.gz
popd

ln -s "${ROOT}" "${MIXO}"

export PATH=/tmp/go/bin:"${PATH}"
export GOPATH="/tmp/go"

glide --version

pushd "${MIXO}"
echo $PWD
glide install
go get -u github.com/golang/lint/golint
go get -u github.com/golang/glog
make test
popd


