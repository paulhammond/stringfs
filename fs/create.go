// Copyright 2014 Paul Hammond.
// This software is licensed under the MIT license, see LICENSE.txt for details.

package fs

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

type zipFS struct {
	fs     http.FileSystem
	zip    *zip.Writer
	logger *log.Logger
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
	if zfs.logger != nil {
		zfs.logger.Printf("Added %s", name)
	}
	return err
}

// Create converts a net/http FileSystem to a byte slice. If a logger is
// provided it is used to log informational messages.
func Create(fs http.FileSystem, l *log.Logger) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	zfs := zipFS{fs, w, l}
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

// CreateString converts a net/http Filsystem into a string. If a logger is
// provided it is used to log informational messages.
func CreateString(fs http.FileSystem, l *log.Logger) (string, error) {
	b, err := Create(fs, l)
	return base64.StdEncoding.EncodeToString(b), err
}
