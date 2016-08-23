#!/bin/bash
#
# knoxite
#     Copyright (c) 2016, Stefan Luecke <glaxx@glaxx.net>
#
#   For license see LICENSE.txt
#
if [[ "$OSTYPE" == "linux-gnu" ]]; then
  curl https://dl.minio.io/server/minio/release/linux-amd64/minio --output minio_test
elif [[ "$OSTYPE" == "darwin"* ]]; then
  curl https://dl.minio.io/server/minio/release/darwin-amd64/minio --output minio_test
fi

chmod +x ./minio_test
mkdir s3test

set MINIO_ACCESS_KEY=USWUXHGYZQYFYFFIT3RE
set MINIO_SECRET_KEY=MOJRH0mkL1IPauahWITSVvyDrQbEEIwljvmxdq03

./minio_test server ./s3test &
