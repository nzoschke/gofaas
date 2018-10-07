#!/bin/bash
set -ex

mkdir -p src/$PKG && cd src/$PKG
run -s "Cloning"  git clone $URL --branch $REF --single-branch .
git reset --hard $SHA

run -s "YAML Linting" yamllint -d "{extends: default, rules: {line-length: {max: 140}, key-ordering: {}}}" template.yml

PKGS=$(go list $PKG/...)
run -s "Linting"  golint -set_exit_status $PKGS
run -s "Vetting"  go vet -v $PKGS
run -s "Making"   make -j handlers-go
run -s "Testing"  go test -v $PKGS
