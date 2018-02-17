#!/bin/bash

mkdir -p src/$PKG && cd src/$PKG
run -s "Cloning"  git clone $URL --branch $REF --single-branch .
git reset --hard $SHA

export PYTHONPATH=.pip
run -s "Linting" python3 .pip/yamllint -d "{extends: default, rules: {line-length: {max: 120}, key-ordering: {}}}" template.yml
