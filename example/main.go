// Example is an example of using stringfs.
package main

// To run this example in development mode:
//    go run ./example/main.go
//
// To compile:
//    stringfs example/assets example/assets.go
//    go build -o /tmp/example ./example

import (
	"fmt"
	"net/http"
)

var fileSystem http.FileSystem = http.Dir("./example/assets")

func main() {
	http.Handle("/", http.FileServer(fileSystem))
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		fmt.Println("ListenAndServe: ", err)
	}
}
