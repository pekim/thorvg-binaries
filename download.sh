#! /usr/bin/env bash
set -eo pipefail

cd artifacts
rm -rf artifacts
mkdir -p artifacts

go run download.go
