package ziphttpfs

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"path"
)

type zipFS struct {
	fs  http.FileSystem
	zip *zip.Writer
}

func (zfs zipFS) addDir(name string) error {
	dir, err := zfs.fs.Open(name)
	if err != nil {
		return err
	}

	files, err := dir.Readdir(0)
	if err != nil {
		return err
	}

	for _, stat := range files {
		fileName := path.Join(name, stat.Name())
		if stat.IsDir() {
			err = zfs.addDir(fileName)
		} else {
			err = zfs.addFile(fileName)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (zfs zipFS) addFile(name string) error {
	file, err := zfs.fs.Open(name)
	if err != nil {
		return err
	}
	zf, err := zfs.zip.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(zf, file)
	return err
}

func Create(fs http.FileSystem) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	zfs := zipFS{fs, w}
	err := zfs.addDir(".")
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
