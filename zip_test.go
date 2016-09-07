package httpzip

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHTTPZip(t *testing.T) {
	data, err := createzip()
	if err != nil {
		t.Fatal(err)
	}
	r := bytes.NewReader(data)
	zipFile, err := zip.NewReader(r, r.Size())
	if err != nil {
		t.Fatal(err)
	}

	zip := New(zipFile)

	data, err = zip.GetData("gopher.txt")
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(data, files["gopher.txt"]) {
		t.Error("bad data reading")
	}
	if _, err := zip.GetData("bad"); err == nil {
		t.Error("get bad data")
	}

	ts := httptest.NewServer(zip)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/readme.txt")
	if err != nil {
		t.Fatal(err)
	}
	data, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if !bytes.Equal(data, files["readme.txt"]) {
		t.Error("bad http data reading")
	}

	resp, err = http.Get(ts.URL + "/bad")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Error("reading bad file ok")
	}

	resp, err = http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	data, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if !bytes.Equal(data, files["index.html"]) {
		t.Error("bad http index data reading")
	}

	resp, err = http.Post(ts.URL+"/bad", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Error("reading bad file ok")
	}

	if zip.GetMimeType() != mimetype {
		t.Error("bad mimetype test")
	}

	zip.Close()

	if _, err := zip.Open("read.me"); err != ErrClosed {
		t.Error("bad closed reading")
	}
}

func TestOpenHTTPZip(t *testing.T) {
	data, err := createzip()
	if err != nil {
		t.Fatal(err)
	}
	tmp, err := ioutil.TempFile("", "httpzip")
	if err != nil {
		t.Fatal(err)
	}
	_, err = tmp.Write(data)
	if err != nil {
		tmp.Close()
		t.Fatal(err)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())
	zip, err := Open(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}
	if err := zip.Close(); err != nil {
		t.Error(err)
	}
	if _, err := Open("bad"); err == nil {
		t.Error("open bad file")
	}
}
