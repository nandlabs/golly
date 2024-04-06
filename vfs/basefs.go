package vfs

import (
	"io"
	"net/url"
	"path"

	"oss.nandlabs.io/golly/ioutils"
)

type BaseVFS struct {
	VFileSystem
}

// Copy copies a file from src to dst. If src is a directory, it will copy all files in the directory to dst.

func (b *BaseVFS) Copy(src, dst *url.URL) (err error) {
	// check if src is a directory if so copy all files in the directory to dst
	var srcFile VFile
	var srcFileInfo VFileInfo
	srcFile, err = b.Open(src)
	if err == nil {
		defer ioutils.CloserFunc(srcFile)
		srcFileInfo, err = srcFile.Info()
		if err == nil {
			if srcFileInfo.IsDir() {
				//Prepare url for destination
				var dstFile VFile
				dstFile, err = b.MkdirAll(dst)
				if err == nil {
					defer ioutils.CloserFunc(dstFile)
					var children []VFile
					children, err = srcFile.ListAll()
					if err == nil {
						for _, child := range children {
							var childInfo VFileInfo
							childInfo, err = child.Info()
							if err == nil {
								var childDst *url.URL
								childDst, err = url.Parse(path.Join(dstFile.Url().String(), childInfo.Name()))
								if err == nil {
									err = b.Copy(child.Url(), childDst)
									if err != nil {
										return
									}
								}
							}
						}
					}
				}
			}
		} else {
			var dstFile VFile
			dstFile, err = b.Create(dst)
			if err == nil {
				defer ioutils.CloserFunc(dstFile)
				_, err = io.Copy(dstFile, srcFile)
			}
		}
	}
	return
}

func (b *BaseVFS) CopyRaw(src, dst string) (err error) {
	var srcUrl, dstUrl *url.URL
	srcUrl, err = url.Parse(src)
	if err == nil {
		dstUrl, err = url.Parse(dst)
		if err == nil {
			err = b.Copy(srcUrl, dstUrl)
		}
	}
	return
}

func (b *BaseVFS) CreateRaw(u string) (file VFile, err error) {
	var fileUrl *url.URL
	fileUrl, err = url.Parse(u)
	if err == nil {
		if err == nil {
			file, err = b.Create(fileUrl)
		}
	}
	return
}

func (b *BaseVFS) Delete(src *url.URL) (err error) {
	var srcFile VFile
	var srfFileInfo VFileInfo

	srcFile, err = b.Open(src)
	if err == nil {
		defer ioutils.CloserFunc(srcFile)
		srfFileInfo, err = srcFile.Info()
		if err == nil {
			if srfFileInfo.IsDir() {
				err = srcFile.DeleteAll()
			} else {
				err = srcFile.Delete()
			}
		}
	}
	return
}

func (b *BaseVFS) DeleteRaw(u string) (err error) {
	var fileUrl *url.URL
	fileUrl, err = url.Parse(u)
	if err == nil {
		err = b.Delete(fileUrl)
	}
	return
}

func (b *BaseVFS) List(src *url.URL) (files []VFile, err error) {
	var srcFile VFile
	srcFile, err = b.Open(src)
	if err == nil {
		defer ioutils.CloserFunc(srcFile)
		files, err = srcFile.ListAll()
	}
	return
}

func (b *BaseVFS) ListRaw(src string) (files []VFile, err error) {
	var fileUrl *url.URL
	fileUrl, err = url.Parse(src)
	if err == nil {
		files, err = b.List(fileUrl)
	}
	return
}

func (b *BaseVFS) MkdirRaw(u string) (vFile VFile, err error) {
	var fileUrl *url.URL
	fileUrl, err = url.Parse(u)
	if err == nil {
		vFile, err = b.Mkdir(fileUrl)
	}
	return
}

func (b *BaseVFS) MkdirAllRaw(u string) (vFile VFile, err error) {
	var fileUrl *url.URL
	fileUrl, err = url.Parse(u)
	if err == nil {
		vFile, err = b.MkdirAll(fileUrl)
	}
	return
}

func (b *BaseVFS) Move(src, dst *url.URL) (err error) {
	err = b.Copy(src, dst)
	if err == nil {
		err = b.Delete(src)
	}
	return
}

func (b *BaseVFS) MoveRaw(src, dst string) (err error) {
	var srcUrl, dstUrl *url.URL
	srcUrl, err = url.Parse(src)
	if err == nil {
		dstUrl, err = url.Parse(dst)
		if err == nil {
			err = b.Move(srcUrl, dstUrl)
		}
	}
	return
}

func (b *BaseVFS) OpenRaw(l string) (file VFile, err error) {
	var u *url.URL
	u, err = url.Parse(l)
	if err == nil {
		file, err = b.Open(u)
	}
	return
}

func (b *BaseVFS) Find(location *url.URL, filter FileFilter) (files []VFile, err error) {
	err = b.Walk(location, func(file VFile) (err error) {
		var filterPass bool
		filterPass, err = filter(file)
		if err == nil && filterPass {
			files = append(files, file)
		}
		return
	})
	return
}

func (b *BaseVFS) Walk(u *url.URL, fn WalkFn) (err error) {
	var src VFile
	var srcFi VFileInfo
	var childInfo VFileInfo
	var children []VFile
	src, err = manager.Open(u)
	if err == nil {
		srcFi, err = src.Info()
		if err == nil {
			if srcFi.IsDir() {
				children, err = src.ListAll()
				if err == nil {
					for _, child := range children {
						childInfo, err = child.Info()
						if err == nil {
							if childInfo.IsDir() {
								err = b.Walk(child.Url(), fn)
							} else {
								err = fn(child)
							}
							if err != nil {
								break
							}
						}
					}
				}
			}

		}
	}
	return
}

func (b *BaseVFS) WalkRaw(raw string, fn WalkFn) (err error) {
	var u *url.URL
	u, err = url.Parse(raw)
	if err == nil {
		err = b.Walk(u, fn)
	}
	return
}

func (b *BaseVFS) DeleteMatching(location *url.URL, filter FileFilter) (err error) {
	var files []VFile
	var fileInfo VFileInfo
	files, err = b.Find(location, filter)
	if err == nil {
		for _, file := range files {
			fileInfo, err = file.Info()
			if err == nil {
				if fileInfo.IsDir() {
					err = file.DeleteAll()
					if err != nil {
						break
					}
				} else {
					err = file.Delete()
					if err != nil {
						break
					}
				}
			} else {
				break
			}
		}
	}
	return
}
