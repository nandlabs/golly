package vfs

import (
	"net/url"
)

type WalkFn func(file VFile) error

type VFileSystem interface {
	//Copy File from one location to another. If the source resolves to a  directory, then all its nested children
	//will be copied. The source and destination can be of different filesystems.
	//Since the FS can be different it cannot guarantee to carry forward any symbolic links/shortcuts from source
	//instead it may try to create regular files/directories for the same even if the source and destination have
	//the same schemes
	Copy(src, dst *url.URL) error
	//CopyRaw is same as Copy except it accepts url as string
	CopyRaw(src, dst string) error
	//Create will create a new file this in the specified url. This is a
	Create(u *url.URL) (VFile, error)
	//CreateRaw is same as Create except it accepts the url as a string
	CreateRaw(raw string) (VFile, error)
	//Delete file . if the src resolves to a directory then all the files  and directories under this will be deleted
	Delete(src *url.URL) error
	//DeleteRaw is same as Delete except that it will accept url as a string
	DeleteRaw(src string) error
	//List function will list all the files if the type is a directory
	List(url *url.URL) ([]VFile, error)
	//ListRaw lists the file in the filesystem for a specific url
	ListRaw(url string) ([]VFile, error)
	//Mkdir will create the directory and will throw an error if exists or has permission issues or unable to create
	Mkdir(u *url.URL) (VFile, error)
	//MkdirRaw same as Mkdir, however it accepts  url as string
	MkdirRaw(u string) (VFile, error)
	//MkdirAll  will create all directories missing in the path
	//If the directory already exists it will not throw error, however if the path resolves to a file instead
	//it should error
	MkdirAll(u *url.URL) (VFile, error)
	//MkdirAllRaw same as MkdirAll, however it accepts  url as string
	MkdirAllRaw(u string) (VFile, error)
	//Move will
	Move(src, dst *url.URL) error
	//MoveRaw same as Move except it accepts url as string
	MoveRaw(src, dst string) error
	//Open a file based on the url of the file
	Open(u *url.URL) (VFile, error)
	// OpenRaw is same as Open function, however it accepts the url as string
	OpenRaw(u string) (VFile, error)
	//Schemes is the list of schemes supported by this file system
	Schemes() []string
	//Walk will walk through each of the files in the directory recursively.
	//If the URL resolves to a file it's not expected to throw an error instead the fn just be invoked with the VFile
	//representing the url once
	Walk(url *url.URL, fn WalkFn) error
	//Find files based on filter only works if the file.IsDir() is true
	Find(location *url.URL, filter FileFilter) ([]VFile, error)
	//WalkRaw is same as Walk except that it will accept the url as a string
	WalkRaw(raw string, fn WalkFn) error
	//DeleteMatching will delete only the files that match the filter.
	//If one of the file deletion fails with an error then it stops processing and returns error
	DeleteMatching(location *url.URL, filter FileFilter) error
}

type Manager interface {
	VFileSystem
	Register(vfs VFileSystem)
	IsSupported(scheme string) bool
}
