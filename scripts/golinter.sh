#!/bin/bash

# Copyright Greg Haskins All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

set -e

# shellcheck source=/dev/null
source "$(cd "$(dirname "$0")" && pwd)/functions.sh"

fabric_dir="$(cd "$(dirname "$0")/.." && pwd)"
source_dirs=()
while IFS=$'\n' read -r source_dir; do
    source_dirs+=("$source_dir")
done < <(go list -f '{{.Dir}}' ./... | sed s,"${fabric_dir}".,,g | cut -f 1 -d / | sort -u)

echo "Checking with gofmt"
OUTPUT="$(gofmt -l -s "${source_dirs[@]}")"
OUTPUT="$(filterExcludedAndGeneratedFiles "$OUTPUT")"
if [ -n "$OUTPUT" ]; then
    echo "The following files contain gofmt errors"
    echo "$OUTPUT"
    echo "The gofmt command 'gofmt -l -s -w' must be run for these files"
    exit 1
fi

echo "Checking with goimports"
OUTPUT="$(goimports -l "${source_dirs[@]}")"
OUTPUT="$(filterExcludedAndGeneratedFiles "$OUTPUT")"
if [ -n "$OUTPUT" ]; then
    echo "The following files contain goimports errors"
    echo "$OUTPUT"
    echo "The goimports command 'goimports -l -w' must be run for these files"
    exit 1
fi

echo "Checking with go vet"
PRINTFUNCS="Debug,Debugf,Print,Printf,Info,Infof,Warning,Warningf,Error,Errorf,Critical,Criticalf,Sprint,Sprintf,Log,Logf,Panic,Panicf,Fatal,Fatalf,Notice,Noticef,Wrap,Wrapf,WithMessage"
OUTPUT="$(go vet -all -printfuncs "$PRINTFUNCS" ./...)"
if [ -n "$OUTPUT" ]; then
    echo "The following files contain go vet errors"
    echo "$OUTPUT"
    #exit 1
fi

exit 0

#echo "Checking with staticcheck"
#OUTPUT="$(staticcheck ./... || true)"
#if [ -n "$OUTPUT" ]; then
#    echo "The following staticcheck issues were flagged"
#    echo "$OUTPUT"
#    exit 1
#fi
