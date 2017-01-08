#!/bin/bash
#
# knoxite
#     Copyright (c) 2017, Christian Muehlhaeuser <muesli@gmail.com>
#
#   For license see LICENSE
#

# create ftp chroot
sudo mkdir -p $HOME/ftp/knoxite-citest
sudo chown -R $USER $HOME/ftp

if [[ "$OSTYPE" == "linux-gnu" ]]; then
    # install pure-ftpd
    sudo apt-get -qq update
    sudo apt-get install -y pure-ftpd

    sudo service pure-ftpd stop

    # change port to 2121
    sudo sh -c "echo ',2121' > /etc/pure-ftpd/conf/Bind"

    # Create a password db from the passwd template
    sudo sh -c "echo 'knoxite:\$1\$ONFUv0U0\$zLjMcFT8W7.mQelSHUp2b1:1000:1000::/home/travis/ftp/./::::::::::::' > /etc/pure-ftpd/pureftpd.passwd"
    sudo pure-pw mkdb

    sudo sh -c "echo '/etc/pure-ftpd/pureftpd.pdb' > /etc/pure-ftpd/conf/PureDB"
    sudo ln -s /etc/pure-ftpd/conf/PureDB /etc/pure-ftpd/auth/PureDB

    sudo service pure-ftpd start
elif [[ "$OSTYPE" == "darwin"* ]]; then
    # install pure-ftpd
    brew install pure-ftpd

    sudo chmod -R a+rwx $HOME/ftp

    # Create a password db from the passwd template
    echo 'knoxite:$7$C6..../....ahHkCWvmlfvbTtW2EUe.lW0b.RYkr5.9wL5pkBphre3$wFzTmcYizq91da1qnypUKGSVa7pvbC3vYvI2ZFsHNQ.:98:20::/Users/travis/ftp/./::::::::::::' > /tmp/pureftpd.passwd
    /usr/local/bin/pure-pw mkdb /tmp/pureftpd.pdb -f /tmp/pureftpd.passwd

    sudo /usr/local/sbin/pure-ftpd -E -B -S 127.0.0.1,2121 --minuid=1 -l puredb:/tmp/pureftpd.pdb
fi
