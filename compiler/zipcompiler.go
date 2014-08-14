package main

import (
	"archive/zip"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("")
	var (
		outFile  string
		mimeType string
	)
	flag.StringVar(&outFile, "out", "archive.zft", "output file name")
	flag.StringVar(&mimeType, "mime", "application/x-archive+zip", "archive mimetype")
	flag.Parse()

	archive, err := os.Create(outFile)
	if err != nil {
		log.Fatalln("Error creating file:", err)
	}
	defer func() {
		if err := archive.Close(); err != nil {
			log.Fatalln("Error closing archive file:", err)
		}
	}()

	zipWriter := zip.NewWriter(archive)
	defer func() {
		if err := zipWriter.Close(); err != nil {
			log.Fatalln("Error closing archive:", err)
		}
	}()

	if mimeType != "" {
		stream, err := zipWriter.CreateHeader(&zip.FileHeader{
			Name:   "mimetype",
			Method: zip.Store,
		})
		if err != nil {
			log.Fatalln("Error creating mimetype description:", err)
		}
		if _, err = io.WriteString(stream, mimeType); err != nil {
			log.Fatalln("Error writing mimetype description:", err)
		}
	}

	for _, arg := range flag.Args() {
		file, err := os.Open(arg)
		if err != nil {
			log.Fatalln("Error reading file:", err)
		}
		defer file.Close()
		stream, err := zipWriter.Create(filepath.ToSlash(arg))
		if err != nil {
			log.Fatalln("Error creating file in archive:", err)
		}
		_, err = io.Copy(stream, file)
		if err != nil {
			log.Fatalln("Error copying file to archive:", err)
		}
		log.Println("Added", arg)
	}
}
