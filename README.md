go-nozzle
====

[![Go Documentation](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/rakutentech/go-nozzle)

`go-nozzle` is a library for Go (golang) for building [CloudFoundry(CF) nozzle](https://docs.cloudfoundry.org/loggregator/architecture.html#nozzles). Nozzle is a program which consume data from the Loggregator Firehose and then select, buffer, and transform data, and forward it to other applications, components like [Apache Kafka](http://kafka.apache.org/) or external services like [Data Dog](https://www.datadoghq.com/).

## Install

To install, use `go get`:

```bash
$ go get github.com/rakutentech/go-nozzle
```

## Documentation

Documentation is available on [GoDoc](http://godoc.org/github.com/rakutentech/go-nozzle). Also you can see the example usage of `go-nozzle` on [example](/example) directory. 

## Author

[Taichi Nakashima](https://github.com/tcnksm)
