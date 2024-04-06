package vfs

import (
	"fmt"
	"net/url"
	"sync"
)

var manager Manager

type fileSystems struct {
	mutex       sync.Mutex
	fileSystems map[string]VFileSystem
}

func (fs *fileSystems) Mkdir(url *url.URL) (file VFile, err error) {
	var vfs VFileSystem
	vfs, err = fs.getFsFor(url)
	if err == nil {
		file, err = vfs.Mkdir(url)
	}
	return
}

func (fs *fileSystems) MkdirRaw(raw string) (file VFile, err error) {
	var u *url.URL
	u, err = url.Parse(raw)
	if err == nil {
		file, err = fs.Mkdir(u)
	}
	return
}

func (fs *fileSystems) MkdirAll(url *url.URL) (file VFile, err error) {
	var vfs VFileSystem
	vfs, err = fs.getFsFor(url)
	if err == nil {
		file, err = vfs.MkdirAll(url)
	}
	return
}

func (fs *fileSystems) MkdirAllRaw(raw string) (file VFile, err error) {
	var u *url.URL
	u, err = url.Parse(raw)
	if err == nil {
		file, err = fs.MkdirAll(u)
	}
	return
}

func (fs *fileSystems) Create(u *url.URL) (file VFile, err error) {
	var vfs VFileSystem
	vfs, err = fs.getFsFor(u)
	if err == nil {
		file, err = vfs.Create(u)
	}
	return
}

func (fs *fileSystems) CreateRaw(raw string) (file VFile, err error) {
	var u *url.URL
	u, err = url.Parse(raw)
	if err == nil {
		file, err = fs.Create(u)
	}
	return
}

func (fs *fileSystems) Copy(src, dst *url.URL) (err error) {
	var vfs VFileSystem
	vfs, err = fs.getFsFor(src)
	if err == nil {
		err = vfs.Copy(src, dst)
	}
	return
}

func (fs *fileSystems) CopyRaw(src, dst string) (err error) {
	var srcUrl, dstUrl *url.URL
	srcUrl, err = url.Parse(src)
	if err == nil {
		dstUrl, err = url.Parse(dst)
		if err == nil {
			err = fs.Copy(srcUrl, dstUrl)
		}
	}
	return
}

//func (fs *fileSystems) CopyAll(src, dst *url.URL) (err error) {
//	var vfs VFileSystem
//	vfs, err = fs.getFsFor(src)
//	if err == nil {
//		err = vfs.CopyAll(src, dst)
//	}
//
//	return
//}

//func (fs *fileSystems) CopyAllRaw(src, dst string) (err error) {
//	var srcUrl, dstUrl *url.URL
//	srcUrl, err = url.Parse(src)
//	if err == nil {
//		dstUrl, err = url.Parse(dst)
//		if err == nil {
//			err = fs.CopyAll(srcUrl, dstUrl)
//		}
//	}
//	return
//}

func (fs *fileSystems) Delete(u *url.URL) (err error) {
	var vfs VFileSystem
	vfs, err = fs.getFsFor(u)
	if err == nil {
		err = vfs.Delete(u)
	}
	return
}

func (fs *fileSystems) DeleteRaw(raw string) (err error) {
	var u *url.URL
	u, err = url.Parse(raw)
	if err == nil {
		err = fs.Delete(u)
	}
	return
}

func (fs *fileSystems) DeleteMatching(url *url.URL, filter FileFilter) (err error) {
	var vfs VFileSystem
	vfs, err = fs.getFsFor(url)
	if err == nil {
		err = vfs.DeleteMatching(url, filter)
	}
	return
}

func (fs *fileSystems) List(url *url.URL) (files []VFile, err error) {
	var vfs VFileSystem
	vfs, err = fs.getFsFor(url)
	if err == nil {
		files, err = vfs.List(url)
	}
	return
}

func (fs *fileSystems) ListRaw(raw string) (files []VFile, err error) {
	var u *url.URL
	u, err = url.Parse(raw)
	if err == nil {
		files, err = fs.List(u)
	}
	return
}

func (fs *fileSystems) Move(src, dst *url.URL) (err error) {
	var vfs VFileSystem
	vfs, err = fs.getFsFor(src)
	if err == nil {
		err = vfs.Move(src, dst)
	}
	return
}

func (fs *fileSystems) MoveRaw(src, dst string) (err error) {
	var srcUrl, dstUrl *url.URL
	srcUrl, err = url.Parse(src)
	if err == nil {
		dstUrl, err = url.Parse(dst)
		if err == nil {
			err = fs.Move(srcUrl, dstUrl)
		}
	}
	return
}

func (fs *fileSystems) Open(url *url.URL) (file VFile, err error) {
	var vfs VFileSystem
	vfs, err = fs.getFsFor(url)
	if err == nil {
		file, err = vfs.Open(url)
	}
	return
}

func (fs *fileSystems) OpenRaw(raw string) (file VFile, err error) {
	var u *url.URL
	u, err = url.Parse(raw)
	if err == nil {
		file, err = fs.Open(u)
	}
	return
}

func (fs *fileSystems) Walk(url *url.URL, fn WalkFn) (err error) {
	var vfs VFileSystem
	vfs, err = fs.getFsFor(url)
	if err == nil {
		err = vfs.Walk(url, fn)
	}
	return
}

func (fs *fileSystems) WalkRaw(raw string, fn WalkFn) (err error) {
	var u *url.URL
	u, err = url.Parse(raw)
	if err == nil {
		err = fs.Walk(u, fn)
	}
	return
}

func (fs *fileSystems) Find(url *url.URL, filter FileFilter) (files []VFile, err error) {
	var vfs VFileSystem
	vfs, err = fs.getFsFor(url)
	if err == nil {
		files, err = vfs.Find(url, filter)
	}
	return
}

func (fs *fileSystems) Schemes() (schemes []string) {
	for k := range fs.fileSystems {
		if k == "" {
			continue
		}
		schemes = append(schemes, k)
	}
	return
}

func (fs *fileSystems) IsSupported(scheme string) (supported bool) {
	_, supported = fs.fileSystems[scheme]
	return
}

func (fs *fileSystems) getFsFor(src *url.URL) (vfs VFileSystem, err error) {
	var ok bool
	vfs, ok = fs.fileSystems[src.Scheme]
	if !ok {
		err = fmt.Errorf("Unsupported scheme %s for in the url %s", src.Scheme, src.String())
	}
	return
}

func init() {
	manager = &fileSystems{}
	localFs := newOsFs()
	manager.Register(localFs)
}

func newOsFs() VFileSystem {
	return &OsFs{BaseVFS: &BaseVFS{VFileSystem: &OsFs{}}}
}

func (fs *fileSystems) Register(vfs VFileSystem) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	for _, s := range vfs.Schemes() {
		if fs.fileSystems == nil {
			fs.fileSystems = make(map[string]VFileSystem)
		}
		fs.fileSystems[s] = vfs
	}
}

func GetManager() Manager {
	return manager
}
