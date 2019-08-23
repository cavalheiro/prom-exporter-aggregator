# Prometheus Exporter Aggregator

A way to expose metrics from multiple Prometheus exporters running on the same host, on a single endpoint.

## Why?

Sometimes you need to run multiple Prometheus exporters on a single machine, resulting in multiple ports having to be exposed, and multiple Prometheus queries hitting that host. 

While the "official" approach to such scenarios is to run an additional Prometheus server on the host and configure it to scrape the local targets, this might be too complex or undesirable for scenarios where a simple metric aggregation could work.

## How it works

It works by querying the endpoints defined on the configuration file, aggregating the metrics, and exposing the consolidated output on a single, configurable port. All queries are run in parallel, and it has a very low footprint. 

Since different exporters running on same host can, in some cases, have common metric names, it is possible to define aliases for each endpoint that will be used to prefix the metric names.

## Configuration 

The configuration file is defined as a YAML map 

```yaml
http://url:port1 : alias1
http://url:port2 : alias2
```

The aliases are optional. If specified, the metrics from each endpoint will be prefixed with the alias. 

## How to run

`$ go run prom-exporter-aggregator.go`

#### Command line options

```
-config
    Path to config file (default "prom-exporter-aggregator.yml")
-port
    Port to listen on (default "9191")
``` 
