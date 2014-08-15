package main

import (
	"archive/zip"
	"flag"
	"github.com/alexzorin/csscompress_go"
	"github.com/dchest/jsmin"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("")
	var (
		outFile     string
		mimeType    string
		cssCompress bool
		jsCompress  bool
	)
	flag.StringVar(&outFile, "out", "archive.zft", "output file name")
	flag.StringVar(&mimeType, "mime", "application/x-archive+zip", "archive mimetype")
	flag.BoolVar(&cssCompress, "csscompress", true, "compress css files")
	flag.BoolVar(&jsCompress, "jscompress", true, "compress javascript files")
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

		if cssCompress && filepath.Ext(arg) == ".css" {
			log.Printf("Compressing css %q...", arg)
			css, err := csscompress.New(file, stream)
			if err != nil {
				log.Fatalln("Error creating css compressor:", err)
			}
			css.Minify()
		} else if jsCompress && filepath.Ext(arg) == ".js" {
			log.Printf("Compressing javascript %q...", arg)
			data, err := ioutil.ReadAll(file)
			if err != nil {
				log.Fatalln("Error reading javascript file:", err)
			}
			data, err = jsmin.Minify(data)
			if err != nil {
				log.Fatalln("Error minifing javascript file:", err)
			}
			stream.Write(data)
		} else {
			_, err = io.Copy(stream, file)
			if err != nil {
				log.Fatalln("Error copying file to archive:", err)
			}
		}
		log.Println("Added", arg)
	}
}
