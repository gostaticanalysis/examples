# externalpkg

This is an example for type checking with a package which is not in dependencies.

```go
$ go build -o externalpkg
$ cd testdata/a
$ ../../externalpkg github.com/tenntenn/goplayground ./...
true
```
