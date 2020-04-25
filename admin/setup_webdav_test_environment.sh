#!/bin/bash
#
# knoxite
#     Copyright (c) 2020, Fabian Siegel <fabians1999@gmail.com>
#
#   For license see LICENSE
#

GO_WEBDAV_VERSION="v3.0.0"
WEBDAV_DIR="$HOME/webdav"
WEBDAV_SUBDIR="backups"
CONFIG_LOCATION_IN="admin/WEBDAV_CONFIG_IN"
CONFIG_LOCATION="/tmp/webdav-cfg.yaml"

if [[ "$OSTYPE" == "linux-gnu" ]]; then
    OS="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS="darwin"
fi

cp ${CONFIG_LOCATION_IN} ${CONFIG_LOCATION}

# create dir
mkdir -p $WEBDAV_DIR
# Push current directory
pushd . 

cd $WEBDAV_DIR
mkdir -p "${WEBDAV_SUBDIR}"

# download webdav
curl -L "https://github.com/hacdias/webdav/releases/download/${GO_WEBDAV_VERSION}/${OS}-amd64-webdav.tar.gz" --output /tmp/webdav.tar.gz
tar -xzf /tmp/webdav.tar.gz webdav

./webdav -c ${CONFIG_LOCATION} &

popd
