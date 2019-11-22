#!/usr/bin/env bash
set -eux

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

$DIR/wait-for-it.sh db:5432 -- echo 'HORSE!!! Database is up'
$DIR/wait-for-it.sh localstack:4572 -- echo 'HORSE!!! localstack.s3 is up'

go run ./cmd/inari-test-data/main.go

go test ./pkg/...
