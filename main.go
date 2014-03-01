package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/paulhammond/stringfs/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func check(err error) {
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func main() {

	var varName = flag.String("var", "FileSystem", "Variable Name")
	var pkgName = flag.String("pkg", "fs", "Package Name")
	flag.Parse()

	logger := log.New(os.Stdout, "", 0)

	args := flag.Args()
	src := args[0]
	dest := args[1]

	dir := http.Dir(src)
	s, err := fs.CreateString(dir, logger)
	check(err)

	code := new(bytes.Buffer)
	fmt.Fprintf(code, "package %s\n\n", *pkgName)
	fmt.Fprintf(code, "import \"github.com/paulhammond/stringfs/fs\"\n\n")
	fmt.Fprintf(code, "func init() {\n\n")
	fmt.Fprintf(code, "\t%s = fs.Must(fs.New(%q))\n", *varName, s)
	fmt.Fprintf(code, "}\n")

	err = os.MkdirAll(filepath.Dir(dest), 0777)
	check(err)

	err = ioutil.WriteFile(dest, code.Bytes(), 0666)
	check(err)
	logger.Printf("Wrote %s", dest)
}
