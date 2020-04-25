#!/usr/bin/env bash

. ./demo-base.sh

# terminal settings
STTY_CONFIG="$(stty -g < /dev/tty)"
COLUMNS=114
ROWS=20
stty rows $ROWS cols $COLUMNS

# start demos
rm -f help.cast tutorial.cast
asciinema rec -c "./demo_help.sh" help.cast
asciinema rec -c "./demo_tutorial.sh" tutorial.cast

# restore terminal
stty "$STTY_CONFIG" < /dev/tty
