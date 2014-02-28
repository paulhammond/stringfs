package ziphttpfs

import (
	"archive/zip"
	"bytes"
	"github.com/paulhammond/fakehttpfs"
	"reflect"
	"testing"
)

func TestCreate(t *testing.T) {

	fs := fakehttpfs.FileSystem(
		fakehttpfs.File("foo.txt", "foo"),
		fakehttpfs.Dir("1",
			fakehttpfs.File("foo.txt", "foo"),
			fakehttpfs.Dir("1",
				fakehttpfs.File("foo.txt", "foo"),
			),
			fakehttpfs.Dir("2",
				fakehttpfs.File("foo.txt", "foo"),
			),
		),
		fakehttpfs.Dir("2",
			fakehttpfs.File("foo.txt", "foo"),
			fakehttpfs.File("bar.txt", "bar"),
		),
	)

	b, err := Create(fs)
	if err != nil {
		t.Fatalf("unexpected error:", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatalf("unexpected error:", err)
	}

	contents := map[string]string{}
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
		contents[f.Name] = string(body)
	}

	expected := map[string]string{
		"foo.txt":     "foo",
		"1/foo.txt":   "foo",
		"1/1/foo.txt": "foo",
		"1/2/foo.txt": "foo",
		"2/foo.txt":   "foo",
		"2/bar.txt":   "bar",
	}
	if !reflect.DeepEqual(contents, expected) {
		t.Errorf("zip file contents don't match\nhave: %#v\nwant: %#v", contents, expected)
	}
}
