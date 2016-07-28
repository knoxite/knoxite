knoxite
=======

knoxite is a data storage, security & backup system.

### It's secure
knoxite uses AES-encryption to safely store your data.
### It's flexible
You can always extend your storage size or move your stored data to another machine.
### It's efficient
knoxite cleverly de-duplicates your data before storing it.
### It's OpenSource
knoxite is free software. Contribute and spread the word!

## Installation

Make sure you have a working Go environment. Follow the [Go install instructions](http://golang.org/doc/install.html).

First of all you need to checkout the source code:

    go get github.com/knoxite/knoxite
    cd $GOPATH/src/github.com/knoxite/knoxite

Now we need to get the required dependencies:

    go get -v

Let's build knoxite:

    cd knoxite
    go build

Run knoxite --help to see a full list of options.

## Getting started

### Initialize a repository
First of all we need to initialize an empty directory (in this case /tmp/knoxite) as a repository:

    ./knoxite -r /tmp/knoxite -p "my_password" repo init

knoxite encrypts all the data in the repository with the supplied password. Be
warned: if you lose this password, you won't be able to access any of your data.

### Initialize a volume
Each repository can contain several volumes, which store our data organized in snapshots. So let's create one:

    ./knoxite -r /tmp/knoxite -p "my_password" volume init "Backups" -d "My system backups"

### List all volumes
Now you can get a list of all volumes stored in this repository:

    ./knoxite -r /tmp/knoxite -p "my_password" volume list

### Storing data in a volume
Run the following command to create a new snapshot and store your home directory in the newly created volume:

    ./knoxite -r /tmp/knoxite -p "my_password" store [volume ID] $HOME -d "Backup of all my data"

### List all snapshots
Now you can get an overview of all snapshots stored in this volume:

    ./knoxite -r /tmp/knoxite -p "my_password" snapshot list [volume ID]

### Show the content of a snapshot
Running the following command lists the entire content of a snapshot:

    ./knoxite -r /tmp/knoxite -p "my_password" ls [snapshot ID]

### Restoring a snapshot
To restore the latest snapshot to /tmp/myhome, run:

    ./knoxite -r /tmp/knoxite -p "my_password" restore [snapshot ID] -t /tmp/myhome

### Mounting a snapshot
You can even mount a snapshot (currently read-only, read-write is work-in-progress):

    ./knoxite -r /tmp/knoxite -p "my_password" mount [snapshot ID] /mnt

### Backup. No more excuses.

## Development

API docs can be found [here](http://godoc.org/github.com/knoxite/knoxite).

Continuous integration: [![Build Status](https://secure.travis-ci.org/knoxite/knoxite.png)](http://travis-ci.org/knoxite/knoxite)
