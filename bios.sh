#!/bin/bash
set -ex

mkdir -p $REPO && cd $REPO
run -s "Cloning"  git clone $URL --branch $REF --single-branch .
git reset --hard $SHA

run -s "YAML Linting" yamllint -d "{extends: default, rules: {line-length: {max: 140}, key-ordering: {}}}" template.yml

run -s "Linting"  golint -set_exit_status ./...
run -s "Vetting"  go vet ./...
run -s "Making"   make -j handlers-go
run -s "Testing"  go test -v ./...
