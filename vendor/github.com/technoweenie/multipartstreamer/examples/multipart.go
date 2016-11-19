package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

func main() {
	defaultPath, _ := os.Getwd()
	defaultFile := filepath.Join(defaultPath, "streamer.go")
	fullpath := flag.String("path", defaultFile, "Path to the include in the multipart data.")
	flag.Parse()

	buffer := bytes.NewBufferString("")
	writer := multipart.NewWriter(buffer)

	fmt.Println("Adding the file to the multipart writer")
	fileWriter, _ := writer.CreateFormFile("file", *fullpath)
	fileData, _ := os.Open(*fullpath)
	io.Copy(fileWriter, fileData)
	writer.Close()

	fmt.Println("Writing the multipart data to a file")
	output, _ := os.Create("multiparttest")
	io.Copy(output, buffer)
}
