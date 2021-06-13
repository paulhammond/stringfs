# stringfs

Stringfs is a go package that encodes a net/http filesystem into a string.

**⚠️This package is deprecated. You should use the
[embed](https://golang.org/pkg/embed/) package instead of this one. It is better
in every way.**

## Usage

Run `go get github.com/paulhammond/stringfs` to install.

To use stringfs, first declare a variable in your source code:

    var fileSystem http.FileSystem = http.Dir("./example/assets")

Then encode the contents of directory into a string using the stringfs
command:

	stringfs example/assets example/assets.go

Full documentation is available through
[godoc](http://godoc.org/github.com/paulhammond/stringfs).

## License

MIT license, see [LICENSE.txt](LICENSE.txt) for details.
