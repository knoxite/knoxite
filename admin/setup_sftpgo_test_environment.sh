#!/bin/bash
#
# knoxite
#     Copyright (c) 2020, Fabian Siegel <fabians1999@gmail.com>
#
#   For license see LICENSE
#


export SFTP_PASSWORD=test
export SFTP_USER=test
export SFTP_PORT=3000
export SFTP_DIR="$HOME"/sftpgo

if [[ "$OSTYPE" == "linux-gnu" ]]; then
    OS="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS="macOS"
fi

curl -L "https://github.com/drakkan/sftpgo/releases/download/0.9.6/sftpgo_0.9.6_${OS}_x86_64.tar.xz" --output /tmp/sftpgo_tar 
tar -xf /tmp/sftpgo_tar sftpgo

mkdir -p $SFTP_DIR
chmod 777 $SFTP_DIR
./sftpgo portable -u $SFTP_USER -p $SFTP_PASSWORD -s $SFTP_PORT -d $SFTP_DIR -g "*" &
mkdir -p "$HOME"/.ssh
while [ ! -f "id_ecdsa.pub" ]; do sleep 1; done # We need to wait until the host key is generated
echo "[localhost]:$SFTP_PORT $(cat id_ecdsa.pub)" >> "$HOME"/.ssh/known_hosts
