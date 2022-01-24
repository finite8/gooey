package register

import (
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

type VirtualFS struct {
	name    string
	path    string
	files   map[string]*VirtualFile
	folders map[string]*VirtualFS
}

func NewVirtualFS(name string) *VirtualFS {
	return &VirtualFS{
		name:    name,
		path:    "/" + name,
		files:   make(map[string]*VirtualFile),
		folders: make(map[string]*VirtualFS),
	}
}

func (vfs *VirtualFS) newChildFS(name string) *VirtualFS {
	child := NewVirtualFS(name)
	child.path = vfs.path + "/" + name
	return child
}

func (vfs *VirtualFS) SetFileString(fpath, text string) GOOEYFile {
	return vfs.SetFileBytes(fpath, []byte(text))
	// var dirParts []string
	// dir, file := path.Split(fpath)
	// target := vfs
	// for dir != "/" && dir != "." && dir != "" {
	// 	curr := path.Base(dir)
	// 	dirParts = append([]string{curr}, dirParts...)
	// 	dir = path.Dir(dir)
	// }
	// for _, d := range dirParts {
	// 	found, ok := target.folders[d]
	// 	if !ok {
	// 		found = vfs.newChildFS(d)
	// 		target.folders[d] = found
	// 	}
	// 	target = found
	// }
	// nf := newVirtualFile(vfs, file, []byte(text))
	// target.files[file] = nf
	// return nf
}

func (vfs *VirtualFS) SetFileBytes(fpath string, data []byte) GOOEYFile {
	var dirParts []string
	dir, file := path.Split(fpath)
	target := vfs
	for dir != "/" && dir != "." && dir != "" {
		curr := path.Base(dir)
		dirParts = append([]string{curr}, dirParts...)
		dir = path.Dir(dir)
	}
	for _, d := range dirParts {
		found, ok := target.folders[d]
		if !ok {
			found = vfs.newChildFS(d)
			target.folders[d] = found
		}
		target = found
	}
	nf := newVirtualFile(vfs, file, data)
	target.files[file] = nf
	return nf
}

type VirtualFile struct {
	owner    *VirtualFS
	fileName string
	Data     []byte
}

func newVirtualFile(owner *VirtualFS, name string, data []byte) *VirtualFile {
	return &VirtualFile{
		owner:    owner,
		fileName: name,
		Data:     data,
	}
}

type GOOEYFile interface {
	Name() string
	FullPath() string
}

// FullPath returns the full path to the file
func (vf *VirtualFile) FullPath() string {
	return vf.owner.path + "/" + vf.fileName
}

func (vf *VirtualFile) Name() string { // base name of the file
	return vf.fileName
}
func (vf *VirtualFile) Size() int64 { // length in bytes for regular files; system-dependent for others {
	return int64(len(vf.Data))
}
func (vf *VirtualFile) Mode() fs.FileMode { // file mode bits {
	return fs.ModeIrregular
}
func (vf *VirtualFile) ModTime() time.Time { // modification time {
	return time.Now()
}
func (vf *VirtualFile) IsDir() bool { // abbreviation for Mode().IsDir() {
	return false
}
func (vf *VirtualFile) Sys() interface{} { // underlying data source (can return nil) {
	return nil
}

func (vf *VirtualFile) Open() http.File {
	r := bytes.NewReader(vf.Data)
	return &VirtualFileStream{
		Reader: r,
		Source: vf,
	}
}

type VirtualFileStream struct {
	*bytes.Reader
	Source *VirtualFile
}

func (s *VirtualFileStream) Close() error {
	return nil
}

func (s *VirtualFileStream) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, nil
}

func (s *VirtualFileStream) Stat() (fs.FileInfo, error) {
	return s.Source, nil
}

var _ http.FileSystem = (*VirtualFS)(nil)

func (vfs VirtualFS) Open(name string) (http.File, error) {

	parts := strings.Split(name, "/")
	searching := map[string]*VirtualFS{
		vfs.name: &vfs,
	}
	target := &vfs
	for len(parts) != 1 {
		curr := parts[0]
		parts = parts[1:]
		if curr == "" {
			continue
		}
		found, ok := searching[curr]
		if !ok {
			return nil, io.EOF
		}
		target = found
		searching = found.folders
	}
	if f, ok := target.files[parts[0]]; ok {
		return f.Open(), nil
	}
	return nil, io.EOF

}
