#!/bin/sh

docker-compose up -d localstack

sleep 2

awslocal s3 mb s3://events.funabashi.co.uk

docker-compose up --build --exit-code-from tests tests

#docker-compose down -v
