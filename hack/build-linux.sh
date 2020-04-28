#!/usr/bin/env bash

#GO111MODULE=on go build $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $(BINDIR)/$(BINNAME) ./cmd/helm
set -e
DIR=$(dirname $0)

go generate ${DIR}/generate.go
GO111MODULE=on go build -o "${DIR}"/../kcfi ./cmd/kcfi
