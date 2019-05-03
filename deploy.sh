#!/bin/sh

AWS_PROFILE=j4y
IMAGE_NAME=inari-micropub
DOCKER_REPO_URL=725941804651.dkr.ecr.eu-central-1.amazonaws.com/inari-micropub:latest
CLUSTER_NAME=inari-cluster

## build + push docker image
$(aws ecr get-login --no-include-email --region eu-central-1 --profile $AWS_PROFILE)
docker build -t $IMAGE_NAME .
docker tag $IMAGE_NAME:latest $DOCKER_REPO_URL
docker push $DOCKER_REPO_URL

## update service
aws --profile=$AWS_PROFILE ecs update-service --force-new-deployment --service $IMAGE_NAME --cluster $CLUSTER_NAME
