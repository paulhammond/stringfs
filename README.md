# stringfs

Stringfs is a go package that encodes a net/http filesystem into a string.

Most web applications require a number of static assets such as images,
CSS stylesheets etc. Stringfs provides a way to compile those assets into a
single application binary for easier deployment.

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
