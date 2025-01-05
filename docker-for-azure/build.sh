#!/usr/bin/env bash

# Run only from docker-for-azure subdir
docker build -t mytkom/alicetraint ../
docker build -t mytkom/alicetraint-for-azure .
docker tag mytkom/alicetraint-for-azure $IMAGE_NAME