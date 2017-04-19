#!/bin/bash

export GOPATH="$(dirname $(readlink -f $0))"
echo "Getting grab..."
go get github.com/cavaliercoder/grab
echo "Getting resty..."
go get github.com/go-resty/resty
echo "Getting osext..."
go get github.com/kardianos/osext
echo "Getting net..."
go get golang.org/x/net
echo "Done!"
