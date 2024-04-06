package vfs

import (
	"io/fs"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"oss.nandlabs.io/golly/errutils"
	"oss.nandlabs.io/golly/fsutils"
)

type OsFile struct {
	*BaseFile
	file     *os.File
	Location *url.URL
	fs       VFileSystem
}

func (o *OsFile) Close() error {
	return o.file.Close()
}

func (o *OsFile) Read(b []byte) (int, error) {
	return o.file.Read(b)
}

func (o *OsFile) Write(b []byte) (int, error) {
	return o.file.Write(b)
}

func (o *OsFile) Seek(offset int64, whence int) (int64, error) {
	return o.file.Seek(offset, whence)
}

func (o *OsFile) ContentType() string {
	return fsutils.LookupContentType(o.Location.Path)
}

func (o *OsFile) ListAll() (files []VFile, err error) {
	manager := GetManager()
	var children []VFile
	err = filepath.WalkDir(o.Location.Path, visit(manager, &children))
	if err == nil {
		files = children
	}
	return
}

func visit(manager Manager, paths *[]VFile) func(string, os.DirEntry, error) (err error) {
	return func(path string, info os.DirEntry, err2 error) (err error) {
		if err2 != nil {
			return
		}
		if !info.IsDir() {
			var child VFile
			child, err = manager.OpenRaw(path)
			*paths = append(*paths, child)
		}
		return
	}
}

func (o *OsFile) Delete() error {
	return os.Remove(o.Location.Path)
}

func (o *OsFile) DeleteAll() error {
	return os.RemoveAll(o.Location.Path)
}

func (o *OsFile) Info() (VFileInfo, error) {
	return o.file.Stat()
}

func (o *OsFile) Parent() (file VFile, err error) {
	var fileInfos []fs.FileInfo
	fileInfos, err = ioutil.ReadDir(o.Location.Path)
	if err == nil {
		for _, info := range fileInfos {
			var f *os.File
			var u *url.URL
			u, _ = o.Location.Parse("../" + info.Name())
			f, err = os.Open(u.Path)
			if err == nil {
				file = &OsFile{
					file:     f,
					Location: u,
				}
			}
		}
	}
	return
}

func (o *OsFile) Url() *url.URL {
	return o.Location
}

func (o *OsFile) AddProperty(name string, value string) error {
	return errutils.FmtError("Unsupported operation AddProperty for scheme")
}

func (o *OsFile) GetProperty(name string) (v string, err error) {
	err = errutils.FmtError("Unsupported operation GetProperty for scheme")
	return
}
