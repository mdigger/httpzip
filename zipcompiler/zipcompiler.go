// Простой упаковщик в ZIP-файл.
//
// Отличия только в том, что при компрессии CSS, JS и HTML-файлов они автоматически минимизируются.
// Ну и, кроме того, позволяет легко указать mimetype-архива. В этом случае самым первым файлом
// в архив будет добавлен файл с именем mimetype и указанным содержимым. Этот файл добавляется
// без компрессии и может быть легко найден в заголовке файла и проверен без предварительного
// открытия и распаковки архива.
//
// Для просмотра списка параметров запустите приложение с параметром -help:
//  $ ./zipcompiler -help
//  Usage of ./zipcompiler:
//	  -mime="application/x-webarchive+zip": archive mimetype
//	  -mincss=true: minifying css files
//	  -minhtml=true: minifying html files
//	  -minjs=true: minifying javascript files
//	  -out="": output file name
package main

import (
	"archive/zip"
	"flag"
	"github.com/dchest/cssmin"
	"github.com/dchest/htmlmin"
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
		outFile      string
		mimeType     string
		cssCompress  bool
		jsCompress   bool
		htmlCompress bool
	)
	flag.StringVar(&outFile, "out", "", "output file name")
	flag.StringVar(&mimeType, "mime", "application/x-webarchive+zip", "archive mimetype")
	flag.BoolVar(&cssCompress, "mincss", true, "minifying css files")
	flag.BoolVar(&jsCompress, "minjs", true, "minifying javascript files")
	flag.BoolVar(&htmlCompress, "minhtml", true, "minifying html files")
	flag.Parse()

	if outFile == "" || flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	archive, err := os.Create(outFile)
	if err != nil {
		log.Fatalln("Error creating file:", err)
	}
	defer func() {
		if err := archive.Close(); err != nil {
			log.Fatalln("Error closing archive file:", err)
		}
	}()
	log.Print("compress to", outFile)

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

		switch ext := filepath.Ext(arg); {
		case cssCompress && ext == ".css":
			log.Print("*", arg)
			data, err := ioutil.ReadAll(file)
			if err != nil {
				log.Fatalln("Error reading css file:", err)
			}
			stream.Write(cssmin.Minify(data))
		case jsCompress && ext == ".js":
			log.Print("*", arg)
			data, err := ioutil.ReadAll(file)
			if err != nil {
				log.Fatalln("Error reading javascript file:", err)
			}
			data, err = jsmin.Minify(data)
			if err != nil {
				log.Fatalln("Error minifing javascript file:", err)
			}
			stream.Write(data)
		case htmlCompress && (ext == ".html" || ext == ".htm"):
			log.Print("*", arg)
			data, err := ioutil.ReadAll(file)
			if err != nil {
				log.Fatalln("Error reading html file:", err)
			}
			data, err = htmlmin.Minify(data, htmlmin.DefaultOptions)
			if err != nil {
				log.Fatalln("Error minifing html file:", err)
			}
			stream.Write(data)
		default:
			log.Print("+", arg)
			_, err = io.Copy(stream, file)
			if err != nil {
				log.Fatalln("Error copying file to archive:", err)
			}
		}
	}
}
