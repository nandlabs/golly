// Package main demonstrates the ioutils package.
package main

import (
	"fmt"

	"oss.nandlabs.io/golly/ioutils"
)

func main() {
	// MIME type lookups
	fmt.Println("Extension .json ->", ioutils.GetMimeFromExt(".json"))
	fmt.Println("Extension .png  ->", ioutils.GetMimeFromExt(".png"))
	fmt.Println("Extension .mp4  ->", ioutils.GetMimeFromExt(".mp4"))

	// Reverse lookup: MIME to extensions
	fmt.Println("text/html exts:", ioutils.GetExtsFromMime("text/html"))

	// MIME type checks
	fmt.Println("image/png is image?", ioutils.IsImageMime("image/png"))
	fmt.Println("audio/mp3 is audio?", ioutils.IsAudioMime("audio/mp3"))
	fmt.Println("video/mp4 is video?", ioutils.IsVideoMime("video/mp4"))

	// Channel utilities
	ch := make(chan int, 1)
	fmt.Println("Channel closed?", ioutils.IsChanClosed(ch))
	ioutils.CloseChannel(ch)
	fmt.Println("Channel closed after CloseChannel?", ioutils.IsChanClosed(ch))

	// Checksum calculation
	calc := ioutils.NewChkSumCalc(ioutils.SHA256)
	sum, err := calc.Calculate("Hello, Golly!")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("SHA256 checksum:", sum)
	}
}
