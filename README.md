# ![knoxite](/assets/knoxite-logo.svg)

[![Build Status](https://github.com/knoxite/knoxite/workflows/build/badge.svg)](https://github.com/knoxite/knoxite/actions)
[![Coverage Status](https://coveralls.io/repos/github/knoxite/knoxite/badge.svg?branch=master)](https://coveralls.io/github/knoxite/knoxite?branch=master)
[![Go ReportCard](https://goreportcard.com/badge/knoxite/knoxite)](https://goreportcard.com/report/knoxite/knoxite)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://pkg.go.dev/github.com/knoxite/knoxite?tab=doc)

knoxite is a secure data storage & backup system.

Join the discussion on IRC in #knoxite (on irc.libera.chat) or our [Gitter chat room](https://gitter.im/knoxite/chat)!

### :lock: It's secure
knoxite uses AES-encryption to safely store your data.
### :muscle: It's flexible
You can always extend your storage size or move your stored data to another machine.
### :rocket: It's efficient
knoxite cleverly de-duplicates your data before storing it and supports multiple compression algorithms.
### :link: It's connected
You can use multiple storage backends, even parallely: local disks, Dropbox, Amazon S3 & [others](https://knoxite.com/docs/storage-backends/).
### :heart: It's OpenSource
knoxite is free software. Contribute and spread the word!

## Installation

Make sure you have a working Go environment. Follow the [Go install instructions](https://golang.org/doc/install.html).

To install knoxite, simply run:

    git clone https://github.com/knoxite/knoxite.git
    cd knoxite
    go build ./cmd/knoxite/

Or use your favourite package manager:

    # Arch Linux (btw)
    yay -S knoxite-git

Run `knoxite --help` to see a full list of options.

## Getting started

### Initialize a repository
First of all we need to initialize an empty directory (in this case /tmp/knoxite) as a repository:

```
$ knoxite -r /tmp/knoxite repo init
Enter password:
Created new repository at /tmp/knoxite
```

knoxite encrypts all the data in the repository with the supplied password. Be
warned: if you lose this password, you won't be able to access any of your data.

### Initialize a volume
Each repository can contain several volumes, which store our data organized in snapshots. So let's create one:

```
$ knoxite -r /tmp/knoxite volume init "Backups" -d "My system backups"
Volume 66e03034 (Name: Backups, Description: My system backups) created
```

### List all volumes
Now you can get a list of all volumes stored in this repository:

```
$ knoxite -r /tmp/knoxite volume list
ID        Name                              Description
----------------------------------------------------------------------------------------------
66e03034  Backups                           My system backups
```

### Storing data in a volume
Run the following command to create a new snapshot and store your home directory in the newly created volume:

```
$ knoxite -r /tmp/knoxite store [volume ID] $HOME -d "Backup of all my data"
document.txt          5.69 MiB / 5.69 MiB [#########################################] 100.00%
other.txt             4.17 MiB / 4.17 MiB [#########################################] 100.00%
...
Snapshot cebc1213 created: 9 files, 8 dirs, 0 symlinks, 0 errors, 1.23 GiB Original Size, 1.23 GiB Storage Size
```

When errors occur while storing individual data-chunks knoxite still tries to
complete the store operation for the remaining chunks. You can toggle this
behaviour to immediately exit on the first erroroneus data-chunk by setting the
`--pedantic` command line flag.

### List all snapshots
Now you can get an overview of all snapshots stored in this volume:

```
$ knoxite -r /tmp/knoxite snapshot list [volume ID]
ID        Date                 Original Size  Storage Size  Description
----------------------------------------------------------------------------------------------
cebc1213  2016-07-29 02:27:15       1.23 GiB      1.23 GiB  Backup of all my data
----------------------------------------------------------------------------------------------
                                    1.23 GiB      1.23 GiB
```

### Show the content of a snapshot
Running the following command lists the entire content of a snapshot:

```
$ knoxite -r /tmp/knoxite ls [snapshot ID]
Perms       User   Group          Size  ModTime              Name
----------------------------------------------------------------------------------------------
-rw-r--r--  user   group      5.69 MiB  2016-07-29 02:06:04  document.txt
-rw-r--r--  user   group      4.17 MiB  2016-07-29 02:05:22  other.txt
...
```

### Show the content of a snapshotted file
With the following command you can also print out the files content to stdout:
```
$ knoxite -r /tmp/knoxite cat [snapshot ID] document.txt
This is the sample text stored in document.txt
```

### Restoring a snapshot
To restore the latest snapshot to /tmp/myhome, run:

```
$ knoxite -r /tmp/knoxite restore [snapshot ID] /tmp/myhome
document.txt          5.69 MiB / 5.69 MiB [#########################################] 100.00%
other.txt             4.17 MiB / 4.17 MiB [#########################################] 100.00%
...
Restore done: 9 files, 8 dirs, 0 symlinks, 0 errors, 1.23 GiB Original Size, 1.23 GiB Storage Size
```

### Cloning a snapshot
Adds target file or directory to an existing snapshot:

```
$ knoxite -r /tmp/knoxite clone [snapshot ID] $HOME
document.txt          5.89 MiB / 5.89 MiB [#########################################] 100.00%
other.txt             5.10 MiB / 5.10 MiB [#########################################] 100.00%
...
Snapshot aefc4591 created: 9 files, 8 dirs, 0 symlinks, 0 errors, 1.34 GiB Original Size, 1.34 GiB Storage Size
```

### Mounting a snapshot
You can even mount a snapshot (currently read-only, read-write is work-in-progress):

```
$ knoxite -r /tmp/knoxite mount [snapshot ID] /mnt
```

### Backup. No more excuses.

## Configuration System
Knoxite comes bundled with a configuration system. You can declare shorthands
for your repositories and provide default values for settings like encryption,
compression or excludes. For more information refer to the documentation on our
[website](https://knoxite.com/docs/configuration-system/) or take a look into
the `knoxite config` command.

## Optional environment variables
Optionally you can set the `KNOXITE_REPOSITORY` and `KNOXITE_PASSWORD` environment
variables to provide default settings for when no options have been passed to `knoxite`.
