#!/bin/bash
#
# knoxite
#     Copyright (c) 2017, Christian Muehlhaeuser <muesli@gmail.com>
#
#   For license see LICENSE
#

# create ftp chroot
sudo mkdir -p $HOME/knoxite-citest
sudo chown -R $USER /$HOME/knoxite-citest

if [[ "$OSTYPE" == "linux-gnu" ]]; then
    # install pure-ftpd
    sudo apt-get -qq update
    sudo apt-get install -y pure-ftpd

    sudo service pure-ftpd stop

    # change port to 2121
    sudo sh -c "echo ',2121' > /etc/pure-ftpd/conf/Bind"

    # Create a password db from the passwd template
    sudo cp admin/pureftpd.passwd /etc/pure-ftpd/
    sudo pure-pw mkdb

    sudo sh -c "echo '/etc/pure-ftpd/pureftpd.pdb' > /etc/pure-ftpd/conf/PureDB"
    sudo ln -s /etc/pure-ftpd/conf/PureDB /etc/pure-ftpd/auth/PureDB

    sudo service pure-ftpd start
elif [[ "$OSTYPE" == "darwin"* ]]; then
    # install pure-ftpd
    brew install pure-ftpd

    # Create a password db from the passwd template
    /usr/local/bin/pure-pw mkdb /tmp/pureftpd.pdb -f admin/pureftpd.osx.passwd
    sudo chmod a+r /tmp/pureftpd.pdb

    /usr/local/sbin/pure-ftpd -l puredb:/tmp/pureftpd.pdb -E -d -B -S 127.0.0.1,2121
fi
