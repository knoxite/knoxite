#!/bin/bash
#
# knoxite
#     Copyright (c) 2016, Stefan Luecke <glaxx@glaxx.net>
#     Copyright (c) 2017, Christian Muehlhaeuser <muesli@gmail.com>
#
#   For license see LICENSE
#

export MINIO_ACCESS_KEY=USWUXHGYZQYFYFFIT3RE
export MINIO_SECRET_KEY=MOJRH0mkL1IPauahWITSVvyDrQbEEIwljvmxdq03

if [[ "$OSTYPE" == "linux-gnu" ]]; then
    curl https://dl.minio.io/server/minio/release/linux-amd64/minio --output minio_test
elif [[ "$OSTYPE" == "darwin"* ]]; then
    curl https://dl.minio.io/server/minio/release/darwin-amd64/minio --output minio_test
fi

chmod +x ./minio_test
mkdir s3test
./minio_test server ./s3test &
