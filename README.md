# httpzip

    import "github.com/mdigger/httpzip"

Основная идея: возможность быстро подключить zip-файл в качестве отдачи
статических файлов через стандартный Go HTTP-сервер. Специально для этого
реализована поддержка функции, аналогичной http.ServeFile.

## Использование

```go
var ErrNotFound = errors.New("file not found")
```
Ошибка, что файл с таким именем в архиве не найден.

#### type ZipServer

```go
type ZipServer struct {
}
```

ZipServer описывает открытый zip-архив, с поддержкой раздачи содержимого через
HTTP-сервер.

#### func  OpenZipServer

```go
func OpenZipServer(filename string) (*ZipServer, error)
```
OpenZipServer открывает файл с zip-архивом и возвращает ссылку на ZipServer.
Если файл с указанным именем не найден или в процессе открытия произошла ошибка,
то она возвращается.

#### func (*ZipServer) CheckMimeType

```go
func (self *ZipServer) CheckMimeType(mimetype string) bool
```
CheckMimeType проверяет, что самый первый файл в архиве имеет имя mimetype и его
содержимое полностью соответствует строке, переданной в качестве параметра.
Данная проверка используется, например, для открытия EPUB-файлов.

#### func (*ZipServer) Close

```go
func (self *ZipServer) Close() error
```
Close закрывает открытый файл с архивом и завершает работу.

#### func (*ZipServer) GetData

```go
func (self *ZipServer) GetData(name string) ([]byte, error)
```
GetData возвращает содержимое файла с указанным именем. Если такого файла в
архиве нет, то возвращается ошибка.

#### func (*ZipServer) GetFile

```go
func (self *ZipServer) GetFile(name string) (io.ReadCloser, error)
```
GetFile возвращает io.ReadCloser интерфейс для чтения содержимого файла из
архива с указанным именем. Если файл с таким именем в архиве не существует, то
возвращается ошибка.

#### func (*ZipServer) ModTime

```go
func (self *ZipServer) ModTime() time.Time
```
ModTime возвращает время модификации файла с архивом. Используется при отдаче
файлов из архива по HTTP в качестве времени модификации, чтобы осуществить
корректное кеширование.

#### func (*ZipServer) ServeFile

```go
func (self *ZipServer) ServeFile(w http.ResponseWriter, r *http.Request, name string)
```
ServeFile позволяет отдать файл с указанным именем в HTTP-поток. В общем, ради
этой функции все и писалось. При отдаче заголовок Last-Modified устанавливается
в значение времени последней модификации файла с архивом. Кроме этого, корректно
возвращается Content-Type в зависимости от расширения файла.

Функция обрабатывает только GET или HEAD запросы. В противном случае будет
возвращена ошибка http.StatusMethodNotAllowed с корректно установленными
HTTP-заголовками.

Если файла с указанным именем в архиве не найдено, то будет возвращена ошибка
http.StatusNotFound.

При обработке запроса обрабатывается заголовок If-Modified-Since и, если файл не
изменился, возвращается http.StatusNotModified.
