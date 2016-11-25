package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xh4n3/ucloud-sdk-go/service/unet"
	"github.com/xh4n3/ucloud-sdk-go/ucloud"
	"github.com/xh4n3/ucloud-sdk-go/ucloud/auth"
	"github.com/xh4n3/unet_exporter/sdk"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"strconv"
)

var (
	ResourceBandwidthUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "resource_bandwidth_usage",
			Help: "Bandwidth usage per resource",
		},
		[]string{"shareBandwidth", "resource"},
	)
	TotalBandwidthUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "total_bandwidth_usage",
			Help: "Bandwidth usage in total",
		},
		[]string{"shareBandwidth"},
	)
	CurrentBandwidth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "current_bandwidth",
			Help: "Current bandwidth",
		},
		[]string{"shareBandwidth"},
	)
	config *sdk.Config
	uNet   *unet.UNet
)

func init() {
	prometheus.MustRegister(ResourceBandwidthUsage)
	prometheus.MustRegister(TotalBandwidthUsage)
	prometheus.MustRegister(CurrentBandwidth)
}

func main() {
	configContent, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Fatalf("Config file not found: %v", err)
	}

	config = &sdk.Config{}
	err = yaml.Unmarshal(configContent, config)
	if err != nil {
		log.Fatalf("cannot unmarshal config file: %v", err)
	}

	uNet = unet.New(&ucloud.Config{
		Credentials: &auth.KeyPair{
			PublicKey:  config.Global.PublicKey,
			PrivateKey: config.Global.PrivateKey,
		},
	})

	shareBandwidth := config.Targets[0]
	collector := sdk.NewCollector(uNet, shareBandwidth)
	collector.ListEIPs()

	go func() {
		for {
			resourceBandwidthMap, bandwidthTotalUsed := collector.ListBandwidthUsages()
			currentBandwidth := collector.GetCurrentBandwidth()

			for resourceName, usage := range resourceBandwidthMap {
				ResourceBandwidthUsage.WithLabelValues(shareBandwidth.Name, resourceName).Set(float64(usage))
			}
			TotalBandwidthUsage.WithLabelValues(shareBandwidth.Name).Set(float64(bandwidthTotalUsed))
			CurrentBandwidth.WithLabelValues(shareBandwidth.Name).Set(float64(currentBandwidth))

			time.Sleep(time.Duration(time.Duration(config.Global.Interval) * time.Second))
		}
	}()

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/trigger", triggerHandler)
	log.Fatal(http.ListenAndServe(config.Global.MertricPort, nil))

}

func triggerHandler(w http.ResponseWriter, req *http.Request) {
	resource := req.URL.Query().Get("resource")
	for _, target := range config.Targets {
		if target.Name == resource {
			up, err := strconv.ParseBool(req.URL.Query().Get("up"))
			break
			err = triggerResizer(up, target)
			if err != nil {
				break
			} else {
				w.WriteHeader(200)
				return
			}
		}
	}
	w.WriteHeader(500)
	return
}

func triggerResizer(up bool, target *sdk.Target) error {
	resizer := sdk.NewResizer(uNet, target)
	var err error
	if up {
		err = resizer.IncreaseBandwidth()
	} else {
		err = resizer.DecreaseBandwidth()
	}
	if err != nil {
		return err
	}
	return nil
}
