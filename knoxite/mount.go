// +build !openbsd
// +build !windows

/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/net/context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"github.com/knoxite/knoxite"
)

// CmdMount describes the command
type CmdMount struct {
	global *GlobalOptions

	ready chan struct{}
	done  chan struct{}

	repository knoxite.Repository
}

func init() {
	_, err := parser.AddCommand("mount",
		"mount a snapshot",
		"The mount command mounts a repository read-only to a given directory",
		&CmdMount{
			global: &globalOpts,
			ready:  make(chan struct{}, 1),
			done:   make(chan struct{}),
		})
	if err != nil {
		panic(err)
	}
}

// Usage describes this command's usage help-text
func (cmd CmdMount) Usage() string {
	return "SNAPSHOT-ID MOUNTPOINT"
}

// Execute this command
func (cmd CmdMount) Execute(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf(TWrongNumArgs, cmd.Usage())
	}
	if cmd.global.Repo == "" {
		return ErrMissingRepoLocation
	}

	repository, err := openRepository(cmd.global.Repo, cmd.global.Password)
	if err != nil {
		return err
	}

	mountpoint := args[1]
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

	_, snapshot, err := repository.FindSnapshot(args[0])
	if err != nil {
		return err
	}
	fmt.Println("Updating index")
	updateIndex(&repository, snapshot)
	fmt.Println("Updating index done")
	for _, arc := range root.Items {
		roottree.Add(arc.Item.Path, arc)
	}

	cmd.ready <- struct{}{}

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
	case <-cmd.done:
		err := fuse.Unmount(mountpoint)
		if err != nil {
			fmt.Printf("Error umounting: %s\n", err)
		}
		return c.Close()
	}
}

// Node in our virtual filesystem
type Node struct {
	Items      map[string]*Node
	Item       knoxite.ItemData
	Repository *knoxite.Repository
	//	sync.RWMutex
}

var (
	root *Node
)

func node(name string) *Node {
	l := strings.Split(name, string(filepath.Separator))

	item := root
	for _, s := range l {
		if len(s) == 0 {
			continue
		}
		fmt.Println("Finding:", s)
		v, ok := item.Items[s]
		if !ok {
			fmt.Println("Adding to tree:", s)
			v = &Node{}
			v.Items = make(map[string]*Node)
			item.Items[s] = v
		}

		item = v
	}

	return item
}

func updateIndex(repository *knoxite.Repository, snapshot *knoxite.Snapshot) {
	root = &Node{}
	root.Items = make(map[string]*Node)
	for _, arc := range snapshot.Items {
		fmt.Println("Adding to index:", arc.Path)
		i := node(arc.Path)
		i.Item = arc
		i.Repository = repository
	}
}

// Attr returns this node's filesystem attr's
func (node *Node) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1
	a.Mode = node.Item.Mode
	a.Size = node.Item.Size

	switch node.Item.Type {
	case knoxite.SymLink:
		a.Mode |= os.ModeSymlink
	}

	return nil
}

// Lookup is used to stat items
func (node *Node) Lookup(ctx context.Context, name string) (fs.Node, error) {
	item, ok := node.Items[name]
	if ok {
		return item, nil
	}

	return nil, fuse.ENOENT
}

// ReadDirAll returns all directories directly below this node
func (node *Node) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	dirDirs := []fuse.Dirent{}

	for k := range node.Items {
		ent := fuse.Dirent{Inode: 2, Name: k, Type: fuse.DT_File}
		dirDirs = append(dirDirs, ent)
	}

	return dirDirs, nil
}

// Open opens a file
func (node *Node) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	if !req.Flags.IsReadOnly() {
		return nil, fuse.Errno(syscall.EACCES)
	}
	resp.Flags |= fuse.OpenKeepCache
	return node, nil
}

// Read reads from a file
func (node *Node) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	d, err := knoxite.ReadArchive(*node.Repository, node.Item, int(req.Offset), req.Size)
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

// Readlink returns the target a symlink is pointing to
func (node *Node) Readlink(ctx context.Context, req *fuse.ReadlinkRequest) (string, error) {
	return node.Item.PointsTo, nil
}

// ReadAll reads an entire archive's content
/*func (node *Node) ReadAll(ctx context.Context) ([]byte, error) {
	d, _, err := knoxite.DecodeArchiveData(*node.Repository, node.Item)
	if err != nil {
		return d, err
	}
	return d, nil
}*/
