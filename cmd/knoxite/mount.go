// +build !openbsd
// +build !windows

/*
 * knoxite
 *     Copyright (c) 2016-2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/knoxite/knoxite"
)

var (
	mountCmd = &cobra.Command{
		Use:   "mount <snapshot> <target>",
		Short: "mount a snapshot",
		Long:  `The mount command mounts a repository read-only to a given directory`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("mount needs to know which snapshot to work on")
			}
			if len(args) < 2 {
				return fmt.Errorf("mount needs to know where to mount the snapshot to")
			}
			return executeMount(args[0], args[1])
		},
	}
)

func init() {
	RootCmd.AddCommand(mountCmd)
}

func executeMount(snapshotID, mountpoint string) error {
	repository, err := openRepository(globalOpts.Repo, globalOpts.Password, false)
	if err != nil {
		return err
	}
	defer repository.Close()

	_, snapshot, err := repository.FindSnapshot(snapshotID)
	if err != nil {
		return err
	}

	if _, serr := os.Stat(mountpoint); os.IsNotExist(serr) {
		fmt.Printf("Mountpoint %s doesn't exist, creating it\n", mountpoint)
		err = os.Mkdir(mountpoint, os.ModeDir|0700)
		if err != nil {
			return err
		}
	}
	c, err := fuse.Mount(
		mountpoint,
		fuse.ReadOnly(),
		fuse.FSName("knoxite"),
	)
	if err != nil {
		return err
	}

	roottree := fs.Tree{}

	fmt.Println("Updating index")
	updateIndex(&repository, snapshot)
	fmt.Println("Updating index done")
	for _, arc := range root.Items {
		roottree.Add(arc.Archive.Path, arc)
	}

	ready := make(chan struct{}, 1)
	done := make(chan struct{})
	ready <- struct{}{}

	errServe := make(chan error)
	go func() {
		err = fs.Serve(c, &roottree)
		if err != nil {
			errServe <- err
		}

		<-c.Ready
		errServe <- c.MountError
	}()

	select {
	case err := <-errServe:
		return err
	case <-done:
		err := fuse.Unmount(mountpoint)
		if err != nil {
			fmt.Printf("Error umounting: %s\n", err)
		}
		return c.Close()
	}
}

// Node in our virtual filesystem.
type Node struct {
	Items      map[string]*Node
	Archive    knoxite.Archive
	Repository *knoxite.Repository
	//	sync.RWMutex
}

var (
	root *Node
)

func node(name string, arc knoxite.Archive, repository *knoxite.Repository) *Node {
	l := strings.Split(name, string(filepath.Separator))

	item := root
	for k, s := range l {
		if len(s) == 0 {
			continue
		}
		// fmt.Println("Finding:", s)
		v, ok := item.Items[s]
		if !ok {
			path := filepath.Join(l[:k+1]...)
			fmt.Println("Adding to tree:", path)
			if name != path {
				// We stored an absolute path and need to fake the parent
				// dirs for the first item in the archive
				arc = knoxite.Archive{
					Type:    knoxite.Directory,
					GID:     arc.GID,
					ModTime: arc.ModTime,
					Mode:    arc.Mode,
					Path:    path,
				}
			}

			v = &Node{}
			v.Items = make(map[string]*Node)
			v.Archive = arc
			v.Repository = repository
			item.Items[s] = v
		}

		item = v
	}

	return item
}

func updateIndex(repository *knoxite.Repository, snapshot *knoxite.Snapshot) {
	root = &Node{}
	root.Items = make(map[string]*Node)
	for _, arc := range snapshot.Archives {
		path := arc.Path
		if path[0] == '/' {
			// This archive contains an absolute path
			// Strip the leading slash for mounting
			path = path[1:]
		}
		fmt.Println("Adding to index:", path)
		node(path, *arc, repository)
	}
}

// Attr returns this node's filesystem attributes.
func (node *Node) Attr(ctx context.Context, a *fuse.Attr) error {
	// fmt.Println("Attr:", node.Item.Path)
	// a.Inode = node.Inode
	a.Mode = node.Archive.Mode
	a.Size = node.Archive.Size

	switch node.Archive.Type {
	case knoxite.SymLink:
		a.Mode |= os.ModeSymlink
	case knoxite.Directory:
		a.Mode |= os.ModeDir
	}

	return nil
}

// Lookup is used to stat items.
func (node *Node) Lookup(_ context.Context, name string) (fs.Node, error) {
	// fmt.Println("Lookup:", name)
	item, ok := node.Items[name]
	if ok {
		return item, nil
	}

	return nil, fuse.ENOENT
}

// ReadDirAll returns all items directly below this node.
func (node *Node) ReadDirAll(_ context.Context) ([]fuse.Dirent, error) {
	// fmt.Println("ReadDirAll:", node.Item.Path)
	entries := []fuse.Dirent{}

	for k, v := range node.Items {
		ent := fuse.Dirent{Name: k}
		switch v.Archive.Type {
		case knoxite.File:
			ent.Type = fuse.DT_File
		case knoxite.Directory:
			ent.Type = fuse.DT_Dir
		case knoxite.SymLink:
			ent.Type = fuse.DT_Link
		}

		entries = append(entries, ent)
	}

	return entries, nil
}

// Open opens a file.
func (node *Node) Open(_ context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	if !req.Flags.IsReadOnly() {
		return nil, fuse.Errno(syscall.EACCES)
	}
	resp.Flags |= fuse.OpenKeepCache
	return node, nil
}

// Read reads from a file.
func (node *Node) Read(_ context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	d, err := knoxite.ReadArchive(*node.Repository, node.Archive, int(req.Offset), req.Size)
	if err != nil {
		if err != io.EOF {
			return err
		}
		resp.Data = nil
	} else {
		resp.Data = *d
	}

	return nil
}

// Readlink returns the target a symlink is pointing to.
func (node *Node) Readlink(_ context.Context, _ *fuse.ReadlinkRequest) (string, error) {
	return node.Archive.PointsTo, nil
}

// ReadAll reads an entire archive's content.
/*func (node *Node) ReadAll(_ context.Context) ([]byte, error) {
	d, _, err := knoxite.DecodeArchiveData(*node.Repository, node.Item)
	if err != nil {
		return d, err
	}
	return d, nil
}*/
