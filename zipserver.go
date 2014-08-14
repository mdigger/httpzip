package zipserver

import (
	"archive/zip"
	"errors"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"time"
)

var ErrNotFound = errors.New("file not found")

type ZipServer struct {
	modtime time.Time
	zipFile *zip.ReadCloser
}

func OpenZipServer(filename string) (*ZipServer, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	r, err := zip.OpenReader(filename)
	if err != nil {
		return nil, err
	}
	return &ZipServer{
		modtime: fi.ModTime(),
		zipFile: r,
	}, nil
}

func (self *ZipServer) GetFile(name string) (io.ReadCloser, error) {
	if file := self.findFile(name); file != nil {
		return file.Open()
	}
	return nil, ErrNotFound
}

func (self *ZipServer) GetData(name string) ([]byte, error) {
	r, err := self.GetFile(name)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}

func (self *ZipServer) Close() error {
	return self.zipFile.Close()
}

func (self *ZipServer) ModTime() time.Time {
	return self.modtime
}

func (self *ZipServer) findFile(name string) *zip.File {
	for _, file := range self.zipFile.File {
		if file.Name == name {
			return file
		}
	}
	return nil
}

func (self *ZipServer) ServeFile(w http.ResponseWriter, r *http.Request, name string) {
	println("get:", name)
	wh := w.Header() // HTTP-заголовки ответа для быстрого доступа
	if r.Method != "GET" && r.Method != "HEAD" {
		wh.Set("Allow", "GET")
		wh.Add("Allow", "HEAD")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	file := self.findFile(name) // Ищем файл в нашем архиве
	if file == nil {
		http.NotFound(w, r)
	}
	// Проверяем время модификации файла
	if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil {
		if self.modtime.Before(t.Add(1 * time.Second)) {
			delete(wh, "Content-Type")
			delete(wh, "Content-Length")
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}
	// Всегда устанавливаем заголовок со временем модификации
	wh.Set("Last-Modified", self.modtime.UTC().Format(http.TimeFormat))
	// checkETag
	// etag := r.Header.Get("Etag")
	// rangeReq := r.Header.Get("Range")
	// TODO: Сделать нормальную проверку
	// TODO: Добавить поддержку диапазонов

	// Вычисляет Content-Type по расширению файла
	mimetype := mime.TypeByExtension(path.Ext(name))
	if mimetype == "" {
		mimetype = "application/octet-stream"
	}
	wh.Set("Content-Type", mimetype)

	if r.Method == "HEAD" {
		w.WriteHeader(http.StatusOK)
		return
	}
	fr, err := file.Open()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.Copy(w, fr)
	fr.Close()

}
