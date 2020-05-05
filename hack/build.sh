#!/usr/bin/env bash

#GO111MODULE=on go build $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $(BINDIR)/$(BINNAME) ./cmd/helm
set -e
DIR=$(dirname $0)

which go-bindata >/dev/null || go get -u github.com/go-bindata/go-bindata/...
go generate ${DIR}/generate.go
GO111MODULE=on go build -o "${DIR}"/../kcfi github.com/codefresh-io/kcfi/cmd/kcfi
