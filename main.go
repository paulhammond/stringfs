package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/paulhammond/stringfs/fs"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
