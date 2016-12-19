#!/usr/bin/env bash
curl -s "http://prometheus.api.server/api/v1/query?query=max_over_time(total_bandwidth_usage\[30m\]%20offset%201h)" | jq '.data.result[0].value[1] | tonumber'
