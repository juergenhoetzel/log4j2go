// +build !windows

package filesystem

import (
	"syscall"
)

type Filesystem syscall.Statfs_t

func New(path string) (*Filesystem, error) {
	stat := syscall.Statfs_t{}
	if err := syscall.Statfs(path, &stat); err != nil {
		return nil,  err
	}
	fs := Filesystem(stat)
	return &fs, nil
}

func (fs *Filesystem) SameFs(path string) bool {
	stat := syscall.Statfs_t{}
	syscall.Statfs(path, &stat)
	return (stat.Fsid == fs.Fsid)
}
