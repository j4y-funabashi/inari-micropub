#!/bin/sh

docker-compose down -v
docker-compose up -d localstack

sleep 2

awslocal s3 mb s3://events.funabashi.co.uk
awslocal s3 mb s3://media.funabashi.co.uk

sleep 2

docker-compose up --build --exit-code-from tests tests
