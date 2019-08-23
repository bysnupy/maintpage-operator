#!/bin/bash

BUILD_IMAGE=quay.io/daein/maintpage-operator
IMAGE_TAG=v0.0.1
BUILD_IMAGE_URL=$BUILD_IMAGE:$IMAGE_TAG

operator-sdk build $BUILD_IMAGE_URL --verbose &&
docker push $BUILD_IMAGE_URL
