#!/bin/bash
#
# knoxite
#     Copyright (c) 2017, Christian Muehlhaeuser <muesli@gmail.com>
#
#   For license see LICENSE
#
if [[ "$OSTYPE" == "linux-gnu" ]]; then
  sudo apt-get -qq update
  sudo apt-get install -y bftpd
elif [[ "$OSTYPE" == "darwin"* ]]; then
fi
