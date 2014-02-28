package ziphttpfs

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"os"
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
			err = zfs.addFile(fileName, stat)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (zfs zipFS) addFile(name string, stat os.FileInfo) error {
	file, err := zfs.fs.Open(name)
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(stat)
	if err != nil {
		return err
	}
	header.Name = name
	zf, err := zfs.zip.CreateHeader(header)
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
