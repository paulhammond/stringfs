package ziphttpfs

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"github.com/paulhammond/fakehttpfs"
	"reflect"
	"testing"
	"time"
)

var feb27 = time.Date(2014, 02, 27, 00, 00, 00, 0, time.UTC)
var feb28 = time.Date(2014, 02, 28, 00, 00, 00, 0, time.UTC)

var fakeFS = fakehttpfs.FileSystem(
	fakehttpfs.File("foo.txt", "foo", feb28),
	fakehttpfs.Dir("1",
		fakehttpfs.File("foo.txt", "foo", feb28),
		fakehttpfs.Dir("1",
			fakehttpfs.File("foo.txt", "foo", feb28),
		),
		fakehttpfs.Dir("2",
			fakehttpfs.File("foo.txt", "foo", feb28),
		),
	),
	fakehttpfs.Dir("2",
		fakehttpfs.File("foo.txt", "foo", feb28),
		fakehttpfs.File("bar.txt", "bar", feb27),
	),
)

type testFile struct {
	body    string
	modTime time.Time
}

func TestCreate(t *testing.T) {
	b, err := Create(fakeFS)
	if err != nil {
		t.Fatalf("unexpected error:", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatalf("unexpected error:", err)
	}

	contents := map[string]testFile{}
	for _, f := range reader.File {
		body := make([]byte, f.UncompressedSize64)
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("unexpected error:", err)
		}
		_, err = rc.Read(body)
		if err != nil {
			t.Fatalf("unexpected error:", err)
		}
		rc.Close()
		contents[f.Name] = testFile{
			body:    string(body),
			modTime: f.ModTime(),
		}
	}

	expected := map[string]testFile{
		"foo.txt":     testFile{"foo", feb28},
		"1/foo.txt":   testFile{"foo", feb28},
		"1/1/foo.txt": testFile{"foo", feb28},
		"1/2/foo.txt": testFile{"foo", feb28},
		"2/foo.txt":   testFile{"foo", feb28},
		"2/bar.txt":   testFile{"bar", feb27},
	}
	if !reflect.DeepEqual(contents, expected) {
		t.Errorf("zip file contents don't match\nhave: %#v\nwant: %#v", contents, expected)
	}
}

func TestCreateString(t *testing.T) {
	str, err := CreateString(fakeFS)
	if err != nil {
		t.Fatalf("unexpected error:", err)
	}
	if str == "" {
		t.Errorf("created empty zip file")
	}
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		t.Errorf("string is not base64 encoded")
	}
	_, err = zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Errorf("string is not a base64 encoded zip file")
	}

}
