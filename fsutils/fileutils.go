package fsutils

import (
	"net/http"
	"os"
	"path/filepath"

	"oss.nandlabs.io/golly/ioutils"
)

// FileExists function will check if the file exists in the specified path and if it is a file indeed
func FileExists(path string) bool {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return !fileInfo.IsDir()
}

// DirExists function will check if the Directory exists in the specified path
func DirExists(path string) bool {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return fileInfo.IsDir()
}

// PathExists  will return a boolean if the file/diretory exists
func PathExists(p string) bool {
	_, err := os.Stat(p)
	return !os.IsNotExist(err)
}

// LookupContentType will lookup ContentType based on the file extension.
// This function will only check based on the name of the file and use the file extension.
func LookupContentType(path string) string {
	val := ioutils.GetMimeFromExt(filepath.Ext(path))
	if val == "" {
		val = "application/octet-stream"
	}
	return val
}

// DetectContentType will detect the content type of a file.
func DetectContentType(path string) (string, error) {
	file, err := os.Open(path)
	if err == nil {
		buffer := make([]byte, 512)
		n, err := file.Read(buffer)
		if err == nil {
			return http.DetectContentType(buffer[:n]), nil
		}
	}
	return "", err

}
