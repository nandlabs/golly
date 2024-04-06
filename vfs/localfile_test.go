package vfs

import (
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"oss.nandlabs.io/golly/ioutils"
)

func GetParsedUrl(input string) (output *url.URL) {
	currentPath, _ := os.Getwd()
	u, _ := url.Parse(input)
	path := currentPath + u.Path
	output, _ = url.Parse(u.Scheme + "://" + path)
	return
}

func GetRawPath(input string) (output string) {
	currentPath, _ := os.Getwd()
	u, _ := url.Parse(input)
	path := currentPath + u.Path
	output = u.Scheme + "://" + path
	return
}

var (
	testManager = GetManager()
)

func init() {
	for {
		testManager = GetManager()
		if testManager == nil {
			time.Sleep(100 * time.Millisecond)
		} else {
			break
		}
	}
}

func TestOsFs_Mkdir(t *testing.T) {
	u := GetRawPath("file:///testdata")
	_, err := testManager.MkdirRaw(u)
	if err != nil {
		t.Errorf("MkdirRaw() error = %v", err)
	}
}

func TestOsFs_MkdirAll(t *testing.T) {
	u := GetRawPath("file:///testdata/raw-folder")
	_, err := testManager.MkdirAllRaw(u)
	if err != nil {
		t.Errorf("MkdirAllRaw() error = %v", err)
	}
}

func TestOsFs_Create(t *testing.T) {
	u := GetRawPath("file:///testdata/testFile.txt")
	createdFile, err := testManager.CreateRaw(u)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	fileInfo, err := createdFile.Info()
	if err != nil {
		t.Errorf("Info() error = %v", err)
	}
	if fileInfo.Name() != "testFile.txt" {
		t.Errorf("Invalid file name")
	}

	contentType := createdFile.ContentType()
	if contentType != ioutils.MimeTextPlain {
		t.Errorf("ContentType() invalid")
	}

	fileUrl := createdFile.Url()
	urlPath := GetParsedUrl("file:///testdata/testFile.txt")
	if fileUrl.Path != urlPath.Path {
		t.Errorf("Invalid file URL, got = %s, want = %s", fileUrl, urlPath)
	}

	err = createdFile.AddProperty("test", "value")
	if err.Error() != "Unsupported operation AddProperty for scheme" {
		t.Errorf("Invalid expected error")
	}

	_, err = createdFile.GetProperty("key")
	if err.Error() != "Unsupported operation GetProperty for scheme" {
		t.Errorf("Invalid expected error")
	}

	d1 := []byte("hello\ngo")
	resp, err := createdFile.Write(d1)
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if resp != len(d1) {
		t.Errorf("Invalid data written")
	}
}

func TestOsFs_Open(t *testing.T) {
	u := GetRawPath("file:///testdata/testFile.txt")
	openedFile, err := testManager.OpenRaw(u)
	defer openedFile.Close()

	b1 := make([]byte, 5)
	resp, err := openedFile.Read(b1)
	if err != nil {
		t.Errorf("Read() error = %v", err)
	}
	if resp != len(b1) {
		t.Errorf("Invalid data read")
	}
	_, err = openedFile.Seek(6, 0)
	if err != nil {
		t.Errorf("Seek() error = %v", err)
	}
	b2 := make([]byte, 2)
	n2, err := openedFile.Read(b2)
	if err != nil {
		t.Errorf("Read() error = %v", err)
	}
	if n2 != len(b2) {
		t.Errorf("invalid seek read")
	}
}

func TestOsFile_Delete(t *testing.T) {
	u := GetParsedUrl("file:///testdata/dummyFile.txt")
	createdFile, err := testManager.Create(u)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	err = createdFile.Delete()
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}
}

func TestBaseVFS_CopyRaw(t *testing.T) {
	src := GetRawPath("file:///testdata/testFile.txt")
	dest := GetRawPath("file:///testdata/testFile-copy.txt")
	err := testManager.CopyRaw(src, dest)
	if err != nil {
		t.Errorf("CopyRaw() error = %v", err)
	}
}

func TestBaseVFS_MoveRaw(t *testing.T) {
	src := GetRawPath("file:///testdata/testFile.txt")
	dest := GetRawPath("file:///testdata/testFile-move.txt")

	err := testManager.MoveRaw(src, dest)
	if err != nil {
		t.Errorf("MoveRaw() error = %v", err)
	}
}

func TestBaseVFS_Find(t *testing.T) {
	u := GetParsedUrl("file:///testdata/filterFile.txt")
	_, err := testManager.Create(u)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	filterFunc := func(createdFile VFile) (result bool, err error) {
		var fileInfo VFileInfo
		result = false
		fileInfo, err = createdFile.Info()
		if err != nil {
			return
		}
		if fileInfo.Name() == "filterFile.txt" {
			result = true
		}
		return
	}
	u2 := GetParsedUrl("file:///testdata")
	files, err := testManager.Find(u2, filterFunc)
	if len(files) != 1 {
		t.Errorf("Files not found = %v", err)
	}
}

func TestOsFile_ListAll(t *testing.T) {
	u := GetParsedUrl("file:///testdata/raw-folder/listFile-2.txt")
	_, _ = testManager.Create(u)

	u = GetParsedUrl("file:///testdata")
	output, err := testManager.List(u)
	if err != nil {
		t.Errorf("Error listing files = %v", err)
	}
	for _, item := range output {
		fileInfo, _ := item.Info()
		fmt.Println(fileInfo.Name())
	}
}

func TestOsFsSingleFile_Copy(t *testing.T) {
	src := GetRawPath("file:///testdata/raw-folder")
	dst := GetRawPath("file:///testdata/raw-folder-copy")
	err := testManager.CopyRaw(src, dst)
	if err != nil {
		t.Errorf("Copy() error = %v", err)
	}
}

func TestOsFs_Delete(t *testing.T) {
	u := GetRawPath("file:///testdata")
	err := testManager.DeleteRaw(u)
	if err != nil {
		t.Errorf("DeleteRaw() error = %v", err)
	}
}
