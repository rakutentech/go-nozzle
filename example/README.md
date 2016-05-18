# go-nozzle example

This is example usage of `go-nozzle`.

To build your own nozzle with `go-nozzle`, all you need to do is writing your own producer, received firehose events and produce them to anywhere you want. In this example, we define very simple producer which produces `ValueMetric` event to stdout (using standar golang `log` pacakge). If it detects `slowConsumerAlert`, it prompts its to stderr.

To run this example, you need CloudFoundry environment with loggregator and access authentication for that firehose endpoint.

## Setup

To run this example, you need to export some environmental variables,

```bash
export DOPPLER_ADDR="wss://doppler.cloudfoundry.net"
export UAA_ADDR="https://uaa.cloudfoundry.net"
export CF_USERNAME="tcnksm"
export CF_PASSWORD="fbfanoibNI11"
```

## Usage

After setup, you can run it like below. You can see metrics in your console.

```bash
$ go run main.go
```

To turn off verification of the certificate, use `-insecure` option.

