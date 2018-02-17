#!/bin/bash

mkdir -p src/$PKG && cd src/$PKG
run -s "Cloning"  git clone $URL --branch $REF --single-branch .
git reset --hard $SHA

export PYTHONPATH=.pip
run -s "YAML Linting" python3 .pip/yamllint -d "{extends: default, rules: {line-length: {max: 120}, key-ordering: {}}}" template.yml

PKGS=$(go list $PKG/...)
run -s "Linting"  golint -set_exit_status $PKGS
run -s "Vetting"  go vet -x $PKGS
run -s "Building" go build -v $PKGS
run -s "Testing"  go test -v $PKGS
