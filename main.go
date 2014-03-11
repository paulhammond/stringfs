// Copyright 2014 Paul Hammond.
// This software is licensed under the MIT license, see LICENSE.txt for details.

/*
Stringfs encodes a net/http FileSystem into a string.

Most web applications require a number of static assets such as images,
CSS stylesheets etc. Stringfs provides a way to compile those assets into a
single application binary for easier deployment.

Usage:
    stringfs [flags] directory outputfile

Stringfs takes a directory and an output file. It encodes all files within the
directory into a string in specified output source file.

The flags are:
    -var
        The variable the encoded string will be assigned to. Defaults to
        "fileSystem"
    -pkg
        The package this file will be part of. If not specified stringfs will
        autodetect the package name by checking other files in the same
        directory.

The file created by stringfs does not declare the variable, it just assigns a
value. You are expected to declare the variable yourself in another file,
giving you control over initialization.

A typical usage is to declare the variable with a fallback value. This allows
you to remove the generated file during development, avoiding an extra
compilation step. For example:

    var assets http.FileSystem = http.Dir("./assets")

Alternatively you can initialize the variable and later check for the nil
value. For example:

	// at the top level
	var assets http.FileSystem

	// inside a function
	if assets == nil {
		// working with precompiled assets
	} else {
		// no precompiled assets
	}
*/
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/paulhammond/stringfs/fs"
)

func check(err error) {
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func findPackage(dir string) string {
	pkgs, err := parser.ParseDir(token.NewFileSet(), dir, nil, parser.PackageClauseOnly)
	check(err)
	if len(pkgs) != 1 {
		fmt.Println("Could not autodetect package name, try -pkg flag")
		os.Exit(1)
	}
	for pkg, _ := range pkgs {
		return pkg
	}
	panic("unreachable")
}

func main() {

	var varName = flag.String("var", "fileSystem", "Variable Name")
	var pkgName = flag.String("pkg", "", "Package Name")
	flag.Parse()

	logger := log.New(os.Stdout, "", 0)

	args := flag.Args()
	src := args[0]
	dest := args[1]
	destDir := filepath.Dir(dest)

	if *pkgName == "" {
		*pkgName = findPackage(destDir)
	}

	s, err := fs.CreateString(http.Dir(src), logger)
	check(err)

	code := new(bytes.Buffer)
	fmt.Fprintf(code, "// generated with `%v`\n\n", strings.Join(os.Args, " "))
	fmt.Fprintf(code, "package %s\n\n", *pkgName)
	fmt.Fprintf(code, "import \"github.com/paulhammond/stringfs/fs\"\n\n")
	fmt.Fprintf(code, "func init() {\n")
	fmt.Fprintf(code, "\t%s = fs.Must(fs.New(%q))\n", *varName, s)
	fmt.Fprintf(code, "}\n")

	err = os.MkdirAll(destDir, 0777)
	check(err)

	err = ioutil.WriteFile(dest, code.Bytes(), 0666)
	check(err)
	logger.Printf("Wrote %s", dest)
}
