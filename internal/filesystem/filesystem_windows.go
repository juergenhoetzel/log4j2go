// +build windows

package filesystem

type Filesystem struct{}

func New(path string) (*Filesystem, error) {
	var fs Filesystem
	return &fs, nil
}

func (fs *Filesystem) SameFs(path string) bool {
	return true
}
