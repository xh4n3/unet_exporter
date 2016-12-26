# unet_exporter
Prometheus Exporter for UCloud ShareBandwidth Package

### Dockerfile Usage
```
FROM xh4n3/unet_exporter
COPY config.yml config.yml
```

### Passing PRIVATE_KEY and PUBLIC_KEY via env

Pass PUBLIC_KEY and PUBLIC_KEY by `-e`, which will overrides keys in config file.

### Prometheus
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
