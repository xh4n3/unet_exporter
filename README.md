# unet_exporter
Prometheus exporter for unet

### Dockerfile Usage
```
FROM xh4n3/unet_exporter
COPY config.yml config.yml
```

### Prometheus Alerts
```
ALERT ShareBandwidthTooHigh
  IF 100 * total_bandwidth_usage / current_bandwidth < 85
  FOR 3m
  ANNOTATIONS {
    summary = "ShareBandwidthTooHigh {{$labels.instance}} {{$labels.shareBandwidth}}",
    description = "{{$labels.instance}} - {{$labels.shareBandwidth}} will decrease bandwidth.",
  }

ALERT ShareBandwidthTooLow
  IF total_bandwidth_usage > current_bandwidth
  FOR 3m
  ANNOTATIONS {
    summary = "ShareBandwidthTooLow {{$labels.shareBandwidth}} {{$labels.instance}}",
    description = "{{$labels.instance}} - {{$labels.shareBandwidth}} will increase bandwidth.",
  }
```
