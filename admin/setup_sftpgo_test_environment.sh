#!/bin/bash
#
# knoxite
#     Copyright (c) 2020, Fabian Siegel <fabians1999@gmail.com>
#
#   For license see LICENSE
#

SFTPGO_VERSION="0.9.6"
SFTP_USER=test
SFTP_PASSWORD=test
SFTP_PORT=3000
SFTP_DIR="$HOME"/sftpgo

if [[ "$OSTYPE" == "linux-gnu" ]]; then
    OS="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS="macOS"
fi

# download sftpgo
curl -L "https://github.com/drakkan/sftpgo/releases/download/${SFTPGO_VERSION}/sftpgo_${SFTPGO_VERSION}_${OS}_x86_64.tar.xz" --output /tmp/sftpgo_tar
tar -xf /tmp/sftpgo_tar sftpgo

# create dirs
mkdir -p "$SFTP_DIR"
mkdir -p "$HOME"/.ssh

# start SFTP server
./sftpgo portable -u "$SFTP_USER" -p "$SFTP_PASSWORD" -s $SFTP_PORT -d "$SFTP_DIR" -g "*" &

# wait for it to generate a host key
while [ ! -f "id_ecdsa.pub" ]; do sleep 1; done
echo "[localhost]:$SFTP_PORT $(cat id_ecdsa.pub)" >> "$HOME"/.ssh/known_hosts
