#!/bin/bash
#
# knoxite
#     Copyright (c) 2017, Christian Muehlhaeuser <muesli@gmail.com>
#
#   For license see LICENSE
#
if [[ "$OSTYPE" == "linux-gnu" ]]; then
  sudo apt-get -qq update
  sudo apt-get install -y pure-ftpd

  sudo service pure-ftpd stop

  sudo groupadd ftpgroup
  sudo useradd -g ftpgroup -d /dev/null -s /etc ftpuser
  sudo chown -R ftpuser:ftpgroup /home/ftpuser

  sudo pure-pw useradd knoxite -u ftpuser -d /home/ftpuser
  sudo pure-pw mkdb

  sudo service pure-ftpd start
elif [[ "$OSTYPE" == "darwin"* ]]; then
  echo "Not supported on OSX"
fi
