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
  brew install pure-ftpd

  sudo mkdir -p $HOME/knoxite-citest
  sudo chown -R $USER /$HOME/knoxite-citest

  echo "ENV:"
  id -u
  id -g

  /usr/local/bin/pure-pw mkdb /tmp/pureftpd.pdb -f admin/pureftpd.osx.passwd
  sudo /usr/local/sbin/pure-ftpd -l puredb:/tmp/pureftpd.pdb -E -O clf:/tmp/pureftpd_transfer.log -B &
fi
