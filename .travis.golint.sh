#!/bin/bash

cd "$(dirname $0)"

go get github.com/golang/lint/golint

# Add subdirectories here as we clean up golint on each.
for subdir in .; do
  pushd $subdir
  if golint | grep .; then
    exit 1
  fi
  popd
done
