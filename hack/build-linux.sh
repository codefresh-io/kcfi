#!/usr/bin/env bash

#GO111MODULE=on go build $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $(BINDIR)/$(BINNAME) ./cmd/helm

DIR=$(dirname $0)

GO111MODULE=on go build -o "${DIR}"/../kcfi ./cmd/kcfi
