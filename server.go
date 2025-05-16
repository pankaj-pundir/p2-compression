// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Hello is a simple hello, world demonstration web server.
//
// It serves version information on /version and answers
// any other request like /name by saying "Hello, name!".
//
// See golang.org/x/example/outyet for a more sophisticated server.
package main

import (
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/debug"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: helloserver [options]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

var (
	greeting = flag.String("g", "Hello", "Greet with `greeting`")
	addr     = flag.String("addr", "localhost:8080", "address to serve")
)

type CompressionRequest struct {
	Filename  string `json:"filename"`
	Algorithm string `json:"algorithm"`
	// Add fields for algorithm-specific parameters here
}

func main() {
	// Parse flags.
	flag.Usage = usage

	flag.Parse()

	// Parse and validate arguments (none).
	args := flag.Args()
	if len(args) != 0 {
		usage()
	}

	// Register handlers.
	// All requests not otherwise mapped with go to greet.
	// Serve index.html for the root URL
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/compress", handleCompress)
	http.HandleFunc("/download", handleDownload)
	http.HandleFunc("/generateRandomFile", handleGenerateRandomFile)

	// Serve static files from the "ui" directory
	http.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("ui"))))
	http.HandleFunc("/version", version)

	log.Printf("serving http://%s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func version(w http.ResponseWriter, r *http.Request) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		http.Error(w, "no build information available", 500)
		return
	}

	fmt.Fprintf(w, "<!DOCTYPE html>\n<pre>\n")
	fmt.Fprintf(w, "%s\n", html.EscapeString(info.String()))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "ui/index.html")
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Create a temporary directory if it doesn't exist
	err = os.MkdirAll("temp", os.ModePerm)
	if err != nil {
		http.Error(w, "Error creating temporary directory", http.StatusInternalServerError)
		return
	}

	// Create a new temporary file
	tempFile, err := os.CreateTemp("temp", "upload-*.bin")
	if err != nil {
		http.Error(w, "Error creating temporary file", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	// Copy the uploaded file content to the temporary file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		return
	}

	// In a real application, you'd return the temporary filename or a unique ID
	// to the client for subsequent compression requests.
	fmt.Fprintf(w, "File uploaded successfully!")
}

func handleCompress(w http.ResponseWriter, r *http.Request) {
	// This handler will be implemented later to handle compression requests
	fmt.Fprintf(w, "Compression initiated (handler not fully implemented yet)")
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	// This handler will be implemented later to serve the compressed file
	fmt.Fprintf(w, "Download handler (not fully implemented yet)")
}

func handleGenerateRandomFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fileSizeStr := r.FormValue("size")
	if fileSizeStr == "" {
		http.Error(w, "Missing file size parameter", http.StatusBadRequest)
		return
	}

	fileSize, err := parseFileSize(fileSizeStr)
	if err != nil {
		http.Error(w, "Invalid file size format", http.StatusBadRequest)
		return
	}

	err = os.MkdirAll("temp", os.ModePerm)
	if err != nil {
		http.Error(w, "Error creating temporary directory", http.StatusInternalServerError)
		return
	}

	tempFile, err := os.CreateTemp("temp", "random-*.txt")
	if err != nil {
		http.Error(w, "Error creating temporary file", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	// TODO: Implement actual random text generation of the specified size
	// For now, just create an empty file of the requested size
	err = tempFile.Truncate(fileSize)
	if err != nil {
		http.Error(w, "Error setting file size", http.StatusInternalServerError)
		return
	}

	// Return the temporary filename to the client
	fmt.Fprintf(w, tempFile.Name())
}

// Helper function to parse file size string (e.g., "1MB", "10KB")
func parseFileSize(sizeStr string) (int64, error) {
	// TODO: Implement robust file size parsing with units (KB, MB, GB)
	// For now, assuming the input is in bytes as an integer string
	var size int64
	_, err := fmt.Sscan(sizeStr, &size)
	if err != nil {
		return 0, err
	}
	return size, nil
}
