knoxite
=======

knoxite is a data storage, security & backup system.

### It's secure
knoxite uses AES-encryption to safely store your data.
### It's flexible
You can always extend your storage size or move your stored data to another machine.
### It's efficient
knoxite cleverly de-duplicates your data before storing it.
### It connects
You can use multiple storage backends, even parallely: local disks, Dropbox, Amazon S3 & others.
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

```
$ ./knoxite -r /tmp/knoxite -p "my_password" repo init
Created new repository at /tmp/knoxite
```

knoxite encrypts all the data in the repository with the supplied password. Be
warned: if you lose this password, you won't be able to access any of your data.

### Initialize a volume
Each repository can contain several volumes, which store our data organized in snapshots. So let's create one:

```
$ ./knoxite -r /tmp/knoxite -p "my_password" volume init "Backups" -d "My system backups"
Volume 66e03034 (Name: Backups, Description: My system backups) created
```

### List all volumes
Now you can get a list of all volumes stored in this repository:

```
$ ./knoxite -r /tmp/knoxite -p "my_password" volume list
ID        Name                              Description                                       
----------------------------------------------------------------------------------------------
66e03034  Backups                           My system backups
```

### Storing data in a volume
Run the following command to create a new snapshot and store your home directory in the newly created volume:

```
$ ./knoxite -r /tmp/knoxite -p "my_password" store [volume ID] $HOME -d "Backup of all my data"
document.txt        5.669 MiB / 5.669 MiB [#############################################] 100.00%
other.txt           4.137 MiB / 4.137 MiB [#############################################] 100.00%
...
Snapshot cebc1213 created: 1337 files, 69 dirs, 0 symlinks, 0 errors, 9.772 GiB Original Size, 9.772 GiB Storage Size
```

### List all snapshots
Now you can get an overview of all snapshots stored in this volume:

```
$ ./knoxite -r /tmp/knoxite -p "my_password" snapshot list [volume ID]
ID        Date                 Original Size  Storage Size  Description                                       
--------------------------------------------------------------------------------------------------------------
cebc1213  2016-07-29 02:27:15      9.772 GiB     9.772 GiB  Backup of all my data                             
--------------------------------------------------------------------------------------------------------------
                                   9.772 GiB     9.772 GiB
```

### Show the content of a snapshot
Running the following command lists the entire content of a snapshot:

```
$ ./knoxite -r /tmp/knoxite -p "my_password" ls [snapshot ID]
Perms       User   Group          Size  ModTime              Name                                              
---------------------------------------------------------------------------------------------------------------
-rwxr-xr-x  user   group     9.756 MiB  2016-07-29 02:06:04  knoxite                                           
-rw-r--r--  user   group     1.309 KiB  2016-07-29 02:05:22  main.go                                           
...
```

### Restoring a snapshot
To restore the latest snapshot to /tmp/myhome, run:

```
$ ./knoxite -r /tmp/knoxite -p "my_password" restore [snapshot ID] /tmp/myhome
Creating directory /tmp/myhome
document.txt        5.669 MiB / 5.669 MiB [#############################################] 100.00%
other.txt           4.137 MiB / 4.137 MiB [#############################################] 100.00%
...
Restore done: 1337 files, 69 dirs, 0 symlinks, 0 errors, 9.772 GiB Original Size, 9.772 GiB Storage Size
```

### Cloning a snapshot
It's easy to clone an existing snapshot, adding files to or updating existing files in it:

```
$ ./knoxite -r /tmp/knoxite -p "my_password" clone [snapshot ID] $HOME
newdocument.txt     2.779 MiB / 2.779 MiB [#############################################] 100.00%
changedfile.txt     6.196 MiB / 6.196 MiB [#############################################] 100.00%
...
Snapshot aefc4591 created: 1337 files, 69 dirs, 0 symlinks, 0 errors, 9.775 GiB Original Size, 9.775 GiB Storage Size
```

### Mounting a snapshot
You can even mount a snapshot (currently read-only, read-write is work-in-progress):

```
$ ./knoxite -r /tmp/knoxite -p "my_password" mount [snapshot ID] /mnt
```

### Multiple data storage backends
Adding another data storage backend to the repository:

```
$ ./knoxite -r /tmp/knoxite -p "my_password" repo add dropbox://dropbox.com/knoxite
```

### Backup. No more excuses.

## Development

API docs can be found [here](http://godoc.org/github.com/knoxite/knoxite).

Continuous integration: [![Build Status](https://secure.travis-ci.org/knoxite/knoxite.png)](http://travis-ci.org/knoxite/knoxite)
