// Package main demonstrates the vfs (Virtual File System) package.
package main

import (
	"fmt"
	"os"

	"oss.nandlabs.io/golly/vfs"
)

func main() {
	// Get the VFS manager (supports local file system by default)
	manager := vfs.GetManager()

	// Check if the "file" scheme is supported
	fmt.Println("Supports 'file' scheme:", manager.IsSupported("file"))

	// Create a file using VFS
	tmpDir := os.TempDir()
	filePath := fmt.Sprintf("file://%s/golly-vfs-example.txt", tmpDir)
	fmt.Println("\nWorking with:", filePath)

	// Create and write to a file
	file, err := manager.CreateRaw(filePath)
	if err != nil {
		fmt.Println("Create error:", err)
		return
	}

	n, err := file.WriteString("Hello from golly VFS!\nLine 2\nLine 3\n")
	if err != nil {
		fmt.Println("Write error:", err)
		return
	}
	fmt.Printf("Wrote %d bytes\n", n)
	file.Close()

	// Open and read the file
	file, err = manager.OpenRaw(filePath)
	if err != nil {
		fmt.Println("Open error:", err)
		return
	}
	content, err := file.AsString()
	if err != nil {
		fmt.Println("Read error:", err)
	} else {
		fmt.Println("File contents:", content)
	}

	// Get file info
	info, err := file.Info()
	if err != nil {
		fmt.Println("Info error:", err)
	} else {
		fmt.Printf("File info: name=%s, size=%d, isDir=%v\n", info.Name(), info.Size(), info.IsDir())
	}
	file.Close()

	// List directory contents
	fmt.Println("\nListing temp directory (first 5 entries):")
	dirFile, err := manager.OpenRaw(fmt.Sprintf("file://%s", tmpDir))
	if err != nil {
		fmt.Println("Open dir error:", err)
	} else {
		files, err := dirFile.ListAll()
		if err != nil {
			fmt.Println("ListAll error:", err)
		} else {
			for i, f := range files {
				if i >= 5 {
					fmt.Printf("  ... and %d more entries\n", len(files)-5)
					break
				}
				fi, _ := f.Info()
				fmt.Printf("  %s (dir=%v)\n", fi.Name(), fi.IsDir())
			}
		}
		dirFile.Close()
	}

	// Clean up
	localPath := fmt.Sprintf("%s/golly-vfs-example.txt", tmpDir)
	os.Remove(localPath)
	fmt.Println("\nCleanup: removed temp file")
}
