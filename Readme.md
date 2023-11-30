# Stagesmarts Prometheus Exporter

Metrics exporter to use with Stagesmarts power distribution units.

## Build

Simply build like any other go binary:

    go build .

## Usage

    Usage of prometheus-stagesmarts-exporter:
      -listen-address l string
        	Address the webserver listen on. (default "127.0.0.1")
      -port p int
        	Port the webserver binds to. (default 9005)


## Prometheus configuration

    scrape_configs:
      - job_name: 'stagesmarts'
        static_configs:
          - targets:
            - 192.168.1.2  # StageSmarts PDU
        metrics_path: /metrics
        relabel_configs:
          - source_labels: [__address__]
            target_label: __param_target
          - source_labels: [__param_target]
            target_label: instance
          - target_label: __address__
            replacement: 127.0.0.1:9005  # The Stagesmarts exporter's real hostname:port.

