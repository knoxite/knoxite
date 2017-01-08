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

  sudo mkdir -p $HOME/knoxite-citest
  sudo chown -R $USER /$HOME/knoxite-citest

  sudo cp admin/pureftpd.passwd /etc/pure-ftpd/
  sudo pure-pw mkdb

  sudo sh -c "echo '/etc/pure-ftpd/pureftpd.pdb' > /etc/pure-ftpd/conf/PureDB"
  sudo ln -s /etc/pure-ftpd/conf/PureDB /etc/pure-ftpd/auth/PureDB

  sudo service pure-ftpd start
elif [[ "$OSTYPE" == "darwin"* ]]; then
  echo "Not supported on OSX"
  unset KNOXITE_FTP_URL
fi
