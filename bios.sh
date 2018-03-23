#!/bin/bash
set -ex

mkdir -p src/$PKG && cd src/$PKG
run -s "Cloning"  git clone $URL --branch $REF --single-branch .
git reset --hard $SHA

export PYTHONPATH=vendor/pip
run -s "YAML Linting" python3 vendor/pip/yamllint -d "{extends: default, rules: {line-length: {max: 140}, key-ordering: {}}}" template.yml

PKGS=$(go list $PKG/...)
run -s "Linting"  golint -set_exit_status $PKGS
run -s "Vetting"  go vet -v $PKGS
run -s "Making"   make -j handlers
run -s "Testing"  go test -v $PKGS
