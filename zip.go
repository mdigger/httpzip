// Package httzip allows you to connect a zip archive to the web server as
// static files handler.
package httpzip

import (
	"archive/zip"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type HTTPZip struct {
	zip    *zip.Reader
	files  map[string]*file // the list of file names in the archive
	closer io.Closer
}

func New(r *zip.Reader) *HTTPZip {
	files := make(map[string]*file)
	for i, zf := range r.File {
		if i == 0 && zf.Name == "mimetype" && zf.Method == zip.Store {
			continue
		}
		files["/"+zf.Name] = &file{zf}
	}
	return &HTTPZip{zip: r, files: files}
}

func Open(filename string) (*HTTPZip, error) {
	z, err := zip.OpenReader(filename)
	if err != nil {
		return nil, err
	}
	zip := New(&z.Reader)
	zip.closer = z
	return zip, nil
}

func (h *HTTPZip) Close() error {
	h.files = nil
	if h.closer != nil {
		return h.closer.Close()
	}
	return nil
}

var ErrClosed = errors.New("httpzip: client closed")

func (h *HTTPZip) Open(name string) (*openedFile, error) {
	if h.files == nil {
		return nil, ErrClosed
	}
	name = filepath.ToSlash(name)
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}
	for _, suffix := range []string{"", "index.html", "index.htm"} {
		if zf, ok := h.files[path.Join(name, suffix)]; ok {
			return zf.Open()
		}
	}
	return nil, &os.PathError{"httpzip: open", name, os.ErrNotExist}
}

// GetData возвращает содержимое файла с указанным именем. Если такого файла в
// архиве нет, то возвращается ошибка.
func (h *HTTPZip) GetData(name string) ([]byte, error) {
	r, err := h.Open(name)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}

func (h *HTTPZip) ServeFile(w http.ResponseWriter, r *http.Request, name string) {
	if r.Method != "GET" && r.Method != "HEAD" {
		w.Header().Set("Allow", "GET")
		w.Header().Add("Allow", "HEAD")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}
	f, err := h.Open(name)
	if os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Etag", f.Etag())
	http.ServeContent(w, r, name, f.ModTime(), f)
	f.Close()
}

// ServeHTTP обеспечивает поддержку интерфейса http.Handler.
func (h *HTTPZip) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ServeFile(w, r, r.URL.Path)
}

func (h *HTTPZip) GetMimeType() string {
	if len(h.zip.File) != 0 {
		if firstFile := h.zip.File[0]; firstFile.Name == "mimetype" {
			if file, err := firstFile.Open(); err == nil {
				data, err := ioutil.ReadAll(file)
				file.Close()
				if err == nil {
					return string(data)
				}
			}
		}
	}
	return ""
}
