#!/usr/bin/env bash
# ----------------------------------------------------------
# PURPOSE

# This is the build script for eris:compilers. It will build the tool
# into docker containers in a reliable and predicatable
# manner.

# ----------------------------------------------------------
# REQUIREMENTS

# docker installed locally

# ----------------------------------------------------------
# USAGE

# build_tool.sh

# ----------------------------------------------------------
# Set defaults
set -e
if [ "$CIRCLE_BRANCH" ]
then
  repo=`pwd`
else
  repo=$GOPATH/src/github.com/eris-ltd/eris-compilers
fi
branch=${CIRCLE_BRANCH:=master}
branch=${branch/-/_}
testimage=${testimage:="quay.io/eris/compilers"}

release_min=$(cat $repo/version/version.go | tail -n 1 | cut -d \  -f 4 | tr -d '"')
release_maj=$(echo $release_min | cut -d . -f 1-2)

# ---------------------------------------------------------------------------
# Go!


if [[ "$branch" = "master" ]]
then
  docker build -t $testimage:latest $repo
  docker tag -f $testimage:latest $testimage:$release_maj
  docker tag -f $testimage:latest $testimage:$release_min
else
  docker build -t $testimage:$release_min $repo
fi