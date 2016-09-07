package httpzip

import (
	"archive/zip"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

// Represents an HTTP file, acts as a bridge between zip.File and http.File.
type file struct {
	*zip.File // zip file
}

// Returns an empty slice of files, directory listing is disabled.
func (f *file) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil // directory listing is disabled.
}

// Stats the file.
func (f *file) Stat() (os.FileInfo, error) {
	return f.File.FileInfo(), nil
}

func (f *file) Etag() string {
	return strconv.FormatUint(uint64(f.File.CRC32), 36)
}

func (f *file) Open() (*openedFile, error) {
	ozf, err := f.File.Open()
	if err != nil {
		return nil, err
	}
	return &openedFile{file: f, ReadCloser: ozf}, nil
}

type openedFile struct {
	*file
	io.ReadCloser
	offset int64
}

func (f *openedFile) Read(p []byte) (n int, err error) {
	n, err = f.ReadCloser.Read(p)
	f.offset += int64(n)
	return
}

func (f *openedFile) Seek(offset int64, whence int) (n int64, err error) {
	var size = f.file.FileInfo().Size()
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = f.offset + offset
	case io.SeekEnd:
		abs = size + offset
	default:
		return 0, errors.New("httpzip.file.Seek: invalid whence")
	}

	switch {
	case abs < 0:
		// log.Printf("! bad seek: %v", abs)
		return 0, errors.New("httpzip.file.Seek: negative position")
	case abs == f.offset:
		// log.Printf("= no seek")
		return abs, nil
	case abs > f.offset:
		_, err = io.CopyN(ioutil.Discard, f, abs-f.offset)
		// log.Printf("> seek forward to %d", n)
		return f.offset, err
	default: // case abs < f.offset
		f.offset = 0
		if err = f.ReadCloser.Close(); err != nil {
			return 0, err
		}
		if f.ReadCloser, err = f.file.File.Open(); err != nil {
			return 0, err
		}
		// log.Printf("* close and seek to %d", abs)
		return io.CopyN(ioutil.Discard, f, abs)
	}
}
