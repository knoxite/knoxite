#!/bin/bash
#
# knoxite
#     Copyright (c) 2020, Christian Muehlhaeuser <muesli@gmail.com>
#
#   For license see LICENSE
#

FTPSERVER_VERSION="0.6.0"
FTPSERVER_BINARY="/tmp/ftpserver"
FTP_DIR="$HOME"/knoxite_ftp

if [[ "$OSTYPE" == "linux-gnu" ]]; then
    OS="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS="darwin"
fi

# download ftpserver
curl -L "https://github.com/fclairamb/ftpserverlib/releases/download/v${FTPSERVER_VERSION}/ftpserver-${OS}-amd64" --output ${FTPSERVER_BINARY}
chmod +x ${FTPSERVER_BINARY}

# create dirs
mkdir -p "$FTP_DIR"

# start FTP server
/tmp/ftpserver -data "$FTP_DIR" &

# wait for FTP server to boot up
sleep 5
