// Основная идея: возможность быстро подключить zip-файл в качестве каталога со статическими
// файлами и отдавать их через стандартный Go HTTP-сервер. Специально для этого и реализована
// поддержка функции, аналогичной http.ServeFile.
//
//  // Открываем файл с архивом
//  zipServer, err := httpzip.Open("static.zip")
//  if err != nil {
//      log.Fatal(err)
//  }
//  defer zipServer.Close() // Атоматически закрываем по окончании
//
//  // Инициализируем обработчик HTTP-запросов
//  http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
//      name := strings.TrimPrefix(r.URL.Path, "/static/")
//      zipServer.ServeFile(w, r, name)
//  })
//
// ServeFile позволяет отдать файл с указанным именем в HTTP-поток: в общем, ради этой функции
// все и писалось.
//
// При отдаче заголовок Last-Modified устанавливается в значение времени последней модификации
// файла с архивом. Кроме этого, корректно возвращается Content-Type в зависимости от расширения
// файла. Функция обрабатывает только GET или HEAD запросы. В противном случае будет
// возвращена ошибка http.StatusMethodNotAllowed с корректно установленными HTTP-заголовками.
// Если файла с указанным именем в архиве не найдено, то будет возвращена ошибка
// http.StatusNotFound. При обработке запроса обрабатывается заголовок If-Modified-Since и,
// если файл не изменился, возвращается http.StatusNotModified.
//
// Ну и последнее замечание: данный класс умеет проверять, что самый первый файл имеет имя
// mimetype и его содержимое совпадает с указанным в параметре вызова функции CheckMimeType.
// В общем-то, это, в первую очередь, сделано под влиянием формата EPUB, где таким образом
// определяется, что архив содержит книгу еще до непосредственно распаковки архива. Здесь это
// является не обязательным: просто бесплатное добавление.
//
// Из мелочей: вы можете легко загрузить содержимое любого файла или открыть его как поток
// на чтение. Но это уже так — приятности, не более.
//
//  // Читаем из него, например, шаблон и разбираем его, если это нужно
//  data, err := zipServer.GetData("templates/default.tmpl")
//  if err != nil {
//      log.Fatal(err)
//  }
//  tmpl, err = template.New("").Parse(string(data))
//  if err != nil {
//      log.Fatal(err)
//  }
//
// Остальное — "читайте мои мемуары", как говорил один мой преподаватель, подразумевая, что
// хорошо бы заглянуть в его методичку.package httpzip
package httpzip

import (
	"archive/zip"
	"errors"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

// Ошибка, что файл с таким именем в архиве не найден.
var ErrNotFound = errors.New("file not found")

// HTTPZip описывает открытый zip-архив, с поддержкой раздачи содержимого через HTTP-сервер.
type HTTPZip struct {
	modtime time.Time
	zipFile *zip.ReadCloser
}

// Open открывает файл с zip-архивом и возвращает ссылку на HTTPZip. Если файл с указанным именем
// не найден или в процессе открытия произошла ошибка, то она возвращается.
func Open(filename string) (*HTTPZip, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	r, err := zip.OpenReader(filename)
	if err != nil {
		return nil, err
	}
	return &HTTPZip{
		modtime: fi.ModTime(),
		zipFile: r,
	}, nil
}

// CheckMimeType проверяет, что самый первый файл в архиве имеет имя mimetype и его содержимое
// полностью соответствует строке, переданной в качестве параметра. Данная проверка используется,
// например, для открытия EPUB-файлов.
func (self *HTTPZip) CheckMimeType(mimetype string) bool {
	if len(self.zipFile.File) == 0 {
		return false
	}
	firstFile := self.zipFile.File[0]
	if firstFile.Name != "mimetype" {
		return false
	}
	file, err := firstFile.Open()
	defer file.Close()
	if err != nil {
		return false
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return false
	}
	return string(data) != mimetype
}

// GetFile возвращает io.ReadCloser интерфейс для чтения содержимого файла из архива с указанным
// именем. Если файл с таким именем в архиве не существует, то возвращается ошибка.
func (self *HTTPZip) GetFile(name string) (io.ReadCloser, error) {
	if file := self.findFile(name); file != nil {
		return file.Open()
	}
	return nil, ErrNotFound
}

// GetData возвращает содержимое файла с указанным именем. Если такого файла в архиве нет, то
// возвращается ошибка.
func (self *HTTPZip) GetData(name string) ([]byte, error) {
	r, err := self.GetFile(name)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}

// Close закрывает открытый файл с архивом и завершает работу.
func (self *HTTPZip) Close() error {
	return self.zipFile.Close()
}

// ModTime возвращает время модификации файла с архивом. Используется при отдаче файлов из архива
// по HTTP в качестве времени модификации, чтобы осуществить корректное кеширование.
func (self *HTTPZip) ModTime() time.Time {
	return self.modtime
}

// findFile возвращает ссылку на файл в архиве с указанным именем. Если такого файла не найдено,
// то возвращается nil.
func (self *HTTPZip) findFile(name string) *zip.File {
	name = filepath.ToSlash(name)
	// Убираем корневой слеш в имени файла, если он там есть
	if len(name) > 0 && name[0] == filepath.Separator {
		name = name[1:]
	}
	for _, file := range self.zipFile.File {
		if file.Name == name {
			return file
		}
	}
	return nil
}

// ServeFile позволяет отдать файл с указанным именем в HTTP-поток. В общем, ради этой функции
// все и писалось. При отдаче заголовок Last-Modified устанавливается в значение времени
// последней модификации файла с архивом. Кроме этого, корректно возвращается Content-Type
// в зависимости от расширения файла.
//
// Функция обрабатывает только GET или HEAD запросы. В противном случае будет возвращена ошибка
// http.StatusMethodNotAllowed с корректно установленными HTTP-заголовками.
//
// Если файла с указанным именем в архиве не найдено, то будет возвращена ошибка http.StatusNotFound.
//
// При обработке запроса обрабатывается заголовок If-Modified-Since и, если файл не изменился,
// возвращается http.StatusNotModified.
func (self *HTTPZip) ServeFile(w http.ResponseWriter, r *http.Request, name string) {
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
	// TODO: устанавливать время из файла (???)
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
