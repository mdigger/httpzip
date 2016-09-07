package httpzip

import (
	"archive/zip"
	"bytes"
	"io"
	"log"
	"strings"
	"testing"
)

var (
	files = map[string][]byte{
		"readme.txt": []byte("This archive contains some text files."),
		"gopher.txt": []byte("Gopher names:\nGeorge\nGeoffrey\nGonzo"),
		"todo.txt":   []byte("Get animal handling licence.\nWrite more examples."),
		"index.html": []byte("<html><h1>Index</h1></html>"),
	}
	mimetype = "test"
)

func createzip() ([]byte, error) {
	var zipbuf = new(bytes.Buffer)
	var w = zip.NewWriter(zipbuf)

	stream, err := w.CreateHeader(&zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store,
	})
	if err != nil {
		return nil, err
	}
	if _, err = io.WriteString(stream, mimetype); err != nil {
		return nil, err
	}

	for name, data := range files {
		f, err := w.Create(name)
		if err != nil {
			return nil, err
		}
		_, err = f.Write(data)
		if err != nil {
			return nil, err
		}
	}
	if err = w.Close(); err != nil {
		return nil, err
	}

	return zipbuf.Bytes(), nil
}

func TestZipFile(t *testing.T) {
	data, err := createzip()
	if err != nil {
		t.Fatal(err)
	}
	r := bytes.NewReader(data)
	zipFile, err := zip.NewReader(r, r.Size())
	if err != nil {
		t.Fatal(err)
	}

	zf := file{zipFile.File[1]}
	if zf.Etag() == "" {
		t.Error("bad Etag")
	}
	stat, _ := zf.Stat()
	if stat == nil {
		t.Error("bad stat info")
	}
	if dir, err := zf.Readdir(100); len(dir) > 0 || err != nil {
		t.Error("bad Readdir")
	}

	// pretty.Println(stat)
	ozf, err := zf.Open()
	if err != nil {
		t.Fatal(err)
	}

	log.Printf("\n# %v\n# Seek start\n# %[1]v\n", strings.Repeat("-", 52))
	for _, offset := range []int64{0, 5, 10, 5, 0, 20} {
		log.Printf("## start: %2v (from %2v to %2v)", offset, ozf.offset, offset)
		_, err := ozf.Seek(offset, io.SeekStart)
		if err != nil {
			t.Error(err)
		}
		if offset != ozf.offset {
			t.Errorf("bad start seek: %2v [%2v]", offset, ozf.offset)
		}
	}

	log.Printf("\n# %v\n# Seek current\n# %[1]v\n", strings.Repeat("-", 52))
	var pos = ozf.offset
	for _, offset := range []int64{0, 1, 2, -3, -4, 5} {
		log.Printf("## current: %2v (from %2v to %2v)", offset, ozf.offset, ozf.offset+offset)
		_, err := ozf.Seek(offset, io.SeekCurrent)
		if err != nil {
			t.Error(err)
		}
		pos += offset
		if pos != ozf.offset {
			t.Errorf("bad current seek: %2v [%2v]", offset, ozf.offset)
		}
	}

	log.Printf("\n# %v\n# Seek end\n# %[1]v\n", strings.Repeat("-", 52))
	var size = ozf.FileInfo().Size()
	for _, offset := range []int64{0, -1, -10, -5, -6, -3} {
		log.Printf("## end: %2v (from %2v to %2v)", offset, ozf.offset, size+offset)
		_, err := ozf.Seek(offset, io.SeekEnd)
		if err != nil {
			t.Error(err)
		}
		if size+offset != ozf.offset {
			t.Errorf("bad end seek: %2v [%2v]", offset, ozf.offset)
		}
	}

	if _, err := ozf.Seek(10, 100); err == nil {
		t.Error("bad unknown seek")
	}

	if _, err := ozf.Seek(-1, io.SeekStart); err == nil {
		t.Error("bad negative position seek")
	}
	if ozf.Etag() == "" {
		t.Error("bad Etag")
	}
	stat, _ = ozf.Stat()
	if stat == nil {
		t.Error("bad stat info")
	}
	if dir, err := ozf.Readdir(100); len(dir) > 0 || err != nil {
		t.Error("bad Readdir")
	}

}
