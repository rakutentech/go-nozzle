go-nozzle
====

[![Go Documentation](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/rakutentech/go-nozzle)

`go-nozzle` is a library for Go (golang) for building [CloudFoundry(CF) nozzle](https://docs.cloudfoundry.org/loggregator/architecture.html#nozzles). Nozzle is a program which consume data from the Loggregator Firehose and then select, buffer, and transform data, and forward it to other applications, components or services. 

## Install

To install, use `go get`:

```bash
$ go get github.com/rakutentech/go-nozzle
```

## Documentation

Documentation is available on [GoDoc](http://godoc.org/github.com/rakutentech/go-nozzle).

## Author

[Taichi Nakashima](https://github.com/tcnksm)
