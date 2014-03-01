package fs

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sort"
	"time"
)

type fileSystem struct {
	files map[string]*zip.File
	dirs  map[string][]string
}

func newFS(r *zip.Reader) http.FileSystem {
	files := map[string]*zip.File{}
	dirs := map[string]map[string]bool{}
	for _, f := range r.File {
		name := path.Clean(f.Name)
		if !f.Mode().IsDir() {
			files[name] = f
		}
		var filename string
		for name != "." {
			name, filename = path.Split(name)
			name = path.Clean(name)
			if filename == "" {
				continue
			}
			if _, ok := dirs[name]; !ok {
				dirs[name] = map[string]bool{}
			}
			if dirs[name][filename] {
				break
			}
			dirs[name][filename] = true
		}
	}

	return fileSystem{files, dirMapsToSlices(dirs)}
}

func New(str string) (http.FileSystem, error) {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(data)
	reader, err := zip.NewReader(r, int64(r.Len()))
	if err != nil {
		return nil, err
	}
	return newFS(reader), nil
}

func (fs fileSystem) Open(name string) (http.File, error) {
	name = path.Clean(name)
	if f, ok := fs.files[name]; ok {
		reader, err := f.Open()
		if err != nil {
			return nil, err
		}
		contents, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		reader.Close()
		return file{f, bytes.NewReader(contents)}, nil
	}
	if _, ok := fs.dirs[name]; ok {
		return &dir{fs, name, 0}, nil
	}
	return nil, os.ErrNotExist
}

func (fs fileSystem) stat(name string) (os.FileInfo, error) {
	if f, ok := fs.files[name]; ok {
		return f.FileHeader.FileInfo(), nil
	}
	if _, ok := fs.dirs[name]; ok {
		return &dir{fs, name, 0}, nil
	}
	return nil, os.ErrNotExist
}

type dir struct {
	fs       fileSystem
	name     string
	position int
}

func (d *dir) Stat() (os.FileInfo, error) {
	return d, nil
}

func (d *dir) Readdir(count int) (r []os.FileInfo, err error) {
	files := d.fs.dirs[d.name]

	if count == 0 {
		r = make([]os.FileInfo, len(files))
		for i, f := range files {
			r[i], err = d.fs.stat(path.Join(d.name, f))
			if err != nil {
				return r[0:i], err
			}
		}
	} else {
		r = make([]os.FileInfo, count)
		for i := 0; i < count; i++ {
			if d.position > len(files)-1 {
				return r[0:i], io.EOF
			}
			r[i], err = d.fs.stat(path.Join(d.name, files[d.position]))
			if err != nil {
				return r[0:i], err
			}
			d.position++
		}
	}
	return r, nil
}

func (d *dir) Read(b []byte) (int, error) {
	return 0, errors.New("Not regular file")
}

func (d *dir) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("Not regular file")
}

func (d *dir) Close() error {
	return nil
}

func (d *dir) Name() string {
	return path.Base(d.name)
}

func (d *dir) Size() int64 {
	return 0
}

func (d *dir) Mode() os.FileMode {
	return 0755 | os.ModeDir
}

func (d *dir) ModTime() time.Time {
	panic("unimplemented")
}

func (d *dir) IsDir() bool {
	return true
}

func (d *dir) Sys() interface{} {
	return nil
}

type file struct {
	z *zip.File
	io.ReadSeeker
}

func (f file) Close() error {
	return nil
}

func (f file) Stat() (os.FileInfo, error) {
	return f.z.FileHeader.FileInfo(), nil
}

func (f file) Readdir(count int) ([]os.FileInfo, error) {
	return nil, errors.New("Not dir")
}

func dirMapsToSlices(dirs map[string]map[string]bool) map[string][]string {
	slices := make(map[string][]string, len(dirs))
	for i, d := range dirs {
		slices[i] = make([]string, len(d))
		j := 0
		for k, _ := range d {
			slices[i][j] = k
			j++
		}
		sort.Strings(slices[i])
	}
	return slices
}
