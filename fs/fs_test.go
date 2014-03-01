package fs

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"testing"
	"time"
)

/*
import (
	"archive/zip"
	"bytes"
	"github.com/paulhammond/fakehttpfs"
	"testing"
	"time"
)
*/

// This string contains the fakeFS filesystem from create_test.go
var testStr = "UEsDBBQACAAAAAAAXEQAAAAAAAAAAAAAAAAHAAAAZm9vLnR4dGZvb1BLBwghZXOMAwAAAAMAAABQSwMEFAAIAAAAAABcRAAAAAAAAAAAAAAAAAsAAAAxLzEvZm9vLnR4dDEvMS9mb29QSwcIFlSgcgcAAAAHAAAAUEsDBBQACAAAAAAAXEQAAAAAAAAAAAAAAAALAAAAMS8yL2Zvby50eHQxLzIvZm9vUEsHCMYuADUHAAAABwAAAFBLAwQUAAgAAAAAAFxEAAAAAAAAAAAAAAAABwAAADEvMy50eHQxLzNQSwcIzg8kdwMAAAADAAAAUEsDBBQACAAAAAAAXEQAAAAAAAAAAAAAAAAHAAAAMS80LnR4dDEvNFBLBwhtmkDpAwAAAAMAAABQSwMEFAAIAAAAAABcRAAAAAAAAAAAAAAAAAcAAAAxLzUudHh0MS81UEsHCPuqR54DAAAAAwAAAFBLAwQUAAgAAAAAAFxEAAAAAAAAAAAAAAAACwAAADIvZmViMjgudHh0Mi9mZWIyOFBLBwibqsHABwAAAAcAAABQSwMEFAAIAAAAAABbRAAAAAAAAAAAAAAAAAsAAAAyL2ZlYjI3LnR4dDIvZmViMjdQSwcICrd+UAcAAAAHAAAAUEsBAhQDFAAIAAAAAABcRCFlc4wDAAAAAwAAAAcAAAAAAAAAAAAAAKSBAAAAAGZvby50eHRQSwECFAMUAAgAAAAAAFxEFlSgcgcAAAAHAAAACwAAAAAAAAAAAAAApIE4AAAAMS8xL2Zvby50eHRQSwECFAMUAAgAAAAAAFxExi4ANQcAAAAHAAAACwAAAAAAAAAAAAAApIF4AAAAMS8yL2Zvby50eHRQSwECFAMUAAgAAAAAAFxEzg8kdwMAAAADAAAABwAAAAAAAAAAAAAApIG4AAAAMS8zLnR4dFBLAQIUAxQACAAAAAAAXERtmkDpAwAAAAMAAAAHAAAAAAAAAAAAAACkgfAAAAAxLzQudHh0UEsBAhQDFAAIAAAAAABcRPuqR54DAAAAAwAAAAcAAAAAAAAAAAAAAKSBKAEAADEvNS50eHRQSwECFAMUAAgAAAAAAFxEm6rBwAcAAAAHAAAACwAAAAAAAAAAAAAApIFgAQAAMi9mZWIyOC50eHRQSwECFAMUAAgAAAAAAFtECrd+UAcAAAAHAAAACwAAAAAAAAAAAAAApIGgAQAAMi9mZWIyNy50eHRQSwUGAAAAAAgACAC4AQAA4AEAAAAA"

func TestFileSystem(t *testing.T) {
	fs, err := New(testStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fileTests := []struct {
		path     string
		contents string
		isDir    bool
		modTime  time.Time
	}{
		{path: "", isDir: true},
		{path: ".", isDir: true},
		{path: "foo.txt", contents: "foo", modTime: feb28},
		{path: "1", isDir: true},
		{path: "1/1", isDir: true},
		{path: "1/1/foo.txt", contents: "1/1/foo", modTime: feb28},
		{path: "1/2", isDir: true},
		{path: "1/2/foo.txt", contents: "1/2/foo", modTime: feb28},
		{path: "1/3.txt", contents: "1/3", modTime: feb28},
		{path: "1/4.txt", contents: "1/4", modTime: feb28},
		{path: "1/5.txt", contents: "1/5", modTime: feb28},
		{path: "2", isDir: true},
		{path: "2/feb28.txt", contents: "2/feb28", modTime: feb28},
		{path: "2/feb27.txt", contents: "2/feb27", modTime: feb27},

		// support trailing slashes
		{path: "1/", isDir: true},
		{path: "1/../foo.txt", contents: "foo", modTime: feb28},
		{path: "1//", isDir: true},
		{path: "1/3.txt/", contents: "1/3", modTime: feb28},
	}

	for _, test := range fileTests {
		file, err := fs.Open(test.path)
		if err != nil {
			t.Errorf("expected %s to not error, got %v", test.path, err)
		}
		if file == nil {
			t.Errorf("expected %s to be a file, got nil", test.path)
			continue
		}
		stat, err := file.Stat()
		if err != nil {
			t.Fatalf("expected stat to not error, got %v", err)
		}

		if test.isDir {
			if !stat.IsDir() {
				t.Errorf("expected %s to be a directory, got %v", test.path, file)
			}
		} else {
			b := new(bytes.Buffer)
			b.ReadFrom(file)
			file.Close()
			if s := b.String(); s != test.contents {
				t.Errorf("expected %s to contain %q, got %q", test.path, test.contents, s)
			}
			if mt := stat.ModTime(); mt != test.modTime {
				t.Errorf("expected %s modtime to be %v, got %v", test.path, test.modTime, mt)
			}
		}
	}
}

func TestOpenErrors(t *testing.T) {
	fs, err := New(testStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errTests := []string{
		"/",
		"oops",
		"../foo.txt",
	}

	for _, path := range errTests {
		file, err := fs.Open(path)
		if file != nil {
			t.Errorf("expecred open(%q) to be nil, got %v", path, file)
		}
		if !os.IsNotExist(err) {
			t.Errorf("expecred open(%q) error to be not exist, got %v", path, err)
		}
	}
}

func TestReaddirAll(t *testing.T) {
	fs, err := New(testStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dir, err := fs.Open("1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fileInfos, err := dir.Readdir(0)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	expected := []string{"1", "2", "3.txt", "4.txt", "5.txt"}
	if n := names(fileInfos); !reflect.DeepEqual(n, expected) {
		t.Errorf("filenames don't match\nhave: %#v\nwant: %#v", n, expected)
	}
}

func TestReaddir(t *testing.T) {
	fs, err := New(testStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dir, err := fs.Open("1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		names []string
		err   error
	}{
		{[]string{"1", "2"}, nil},
		{[]string{"3.txt", "4.txt"}, nil},
		{[]string{"5.txt"}, io.EOF},
		{[]string{}, io.EOF},
		{[]string{}, io.EOF},
	}
	for i, test := range tests {
		fileInfos, err := dir.Readdir(2)
		if err != test.err {
			t.Errorf("got error %v, expected %v", err, test.err)
		}
		if n := names(fileInfos); !reflect.DeepEqual(n, test.names) {
			t.Errorf("iteration %d: filenames don't match\nhave: %#v\nwant: %#v", i, n, test.names)
		}
	}
}

func names(infos []os.FileInfo) []string {
	names := make([]string, len(infos))
	for i, info := range infos {
		names[i] = info.Name()
	}
	return names
}
