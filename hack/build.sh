#!/usr/bin/env bash

#GO111MODULE=on go build $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $(BINDIR)/$(BINNAME) ./cmd/helm
set -e
DIR=$(dirname $0)

git submodule update --init
cp -r cf-k8s-agent/k8s-agent charts/
git checkout stage/k8s-agent/config.yaml
cat charts/k8s-agent/values.yaml >> stage/k8s-agent/config.yaml 

which go-bindata >/dev/null || go get -u github.com/go-bindata/go-bindata/...
go generate ${DIR}/generate.go

if [[ -z $CI ]]; then
    GO111MODULE=on go build -o "${DIR}"/../kcfi github.com/codefresh-io/kcfi/cmd/kcfi
else
    echo "Build is running within a CF pipeline, skipping code compilation for goreleaser"
fi
