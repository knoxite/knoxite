#!/usr/bin/env bash

. ./demo-base.sh

# prepare
DEMO_REPO=/tmp/knoxite.demo
rm -Rf "$DEMO_REPO"
mkdir -p "$DEMO_REPO"
cd "$DEMO_REPO"

mkdir data
head -c 1M </dev/urandom >data/important.txt
head -c 1M </dev/urandom >data/stuff.txt

# demo
pe "mkdir backup"
pe "knoxite -r backup/ --password password repo init"
pe "export KNOXITE_REPOSITORY=backup/"
pe "export KNOXITE_PASSWORD=password"
pe "knoxite volume init \"Backups\" -d \"System backups\""
pe "knoxite volume list"
pe "knoxite store latest data/ -d \"Backup of my data\""
pe "knoxite snapshot list latest"
pe "knoxite ls latest"
pe "mkdir restore"
pe "knoxite restore latest restore/"
p ""

# cleanup
rm -Rf "$DEMO_REPO"
