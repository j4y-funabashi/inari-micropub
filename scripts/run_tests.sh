#!/usr/bin/env bash
set -eu

docker-compose down -v
docker-compose up --build -d app

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

wait-for-url() {
    echo "Waiting for ${1}..."
    timeout -s TERM 45 bash -c \
    'while [[ "$(curl -s -o /dev/null -L -w ''%{http_code}'' ${0})" != "200" ]];
    do
        sleep 1;
    done' "${1}"
    echo "HORSE!! ${1} is up"
}

HOST="http://localhost:3040"
wait-for-url ${HOST}

echo "----------"
echo "RUNNING TESTS"
echo "----------"

set +e
go test ./test/...
exit_code="${?}"
set -e

echo "tests exited ${exit_code}"
exit "${exit_code}"
