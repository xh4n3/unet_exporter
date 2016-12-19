package main

import (
	"encoding/json"
	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xh4n3/ucloud-sdk-go/service/umon"
	"github.com/xh4n3/ucloud-sdk-go/service/unet"
	"github.com/xh4n3/ucloud-sdk-go/ucloud"
	"github.com/xh4n3/ucloud-sdk-go/ucloud/auth"
	"github.com/xh4n3/unet_exporter/sdk"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
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
	BandwidthMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bandwidth_metric_up",
			Help: "Bandwidth metric up",
		},
		[]string{"shareBandwidth"},
	)
	config *sdk.Config
	uNet   *unet.UNet
	uMon   *umon.UMon
)

func init() {
	prometheus.MustRegister(ResourceBandwidthUsage)
	prometheus.MustRegister(TotalBandwidthUsage)
	prometheus.MustRegister(CurrentBandwidth)
	prometheus.MustRegister(BandwidthMetric)
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	configContent, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Fatalf("Config file not found: %v", err)
	}

	config = &sdk.Config{}
	err = yaml.Unmarshal(configContent, config)
	if err != nil {
		log.Fatalf("cannot unmarshal config file: %v", err)
	}

	uNet, uMon := envClient(config)

	shareBandwidth := config.Targets[0]
	collector := sdk.NewCollector(uNet, uMon, shareBandwidth)
	collector.ListEIPs()

	log.Println("Collect looping started.")

	go func() {
		for {

			resourceBandwidthMap := collector.ListBandwidthUsages()
			currentBandwidth, err := collector.GetCurrentBandwidth()
			bandwidthTotalUsed := collector.GetTotalBandwidth()

			if err != nil {
				log.Println(err)
				BandwidthMetric.WithLabelValues(shareBandwidth.Name).Set(float64(0))
			} else {
				BandwidthMetric.WithLabelValues(shareBandwidth.Name).Set(float64(1))
			}

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
	go func() {
		err := http.ListenAndServe(config.Global.MertricPort, nil)
		if err != nil {
			switchToDefault()
			log.Fatal(err)
		}
	}()

	//catch exit signals
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)
	switchToDefault()

}

//switch to default bandwidth before exit
func switchToDefault() {
	log.Println("switching to defaultBandwidth before exits")
	for _, target := range config.Targets {
	resizer := sdk.NewResizer(uNet, target, config)
	err := resizer.SetCurrentBandwidth(target.DefaultBandwidth)
	if err != nil {
		log.Printf("switching err: %v", err)
	}
	}
	log.Println("Swiching to default finished, exiting.")
}

func triggerHandler(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)

	var webhookMessage notify.WebhookMessage
	err := decoder.Decode(&webhookMessage)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	var shareBandwidthID string
	var up bool
	isResizerWebhook := false
	for _, alert := range webhookMessage.Alerts {
		if alert.Labels["job"] == "bandwidth-resizer" {
			isResizerWebhook = true
			shareBandwidthID = alert.Labels["shareBandwidth"]
			if alert.Labels["alertname"] == "ShareBandwidthTooLow" {
				up = true
			}
			break
		}
	}
	if !isResizerWebhook {
		return
	}

	for _, target := range config.Targets {
		if target.Name == shareBandwidthID {
			err = triggerResizer(up, target)
			if err != nil {
				log.Println(err)
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
	resizer := sdk.NewResizer(uNet, target, config)
	var err error
	if up {
		err = resizer.IncreaseBandwidth()
	} else {
		err = resizer.DecreaseBandwidth()
	}
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func envClient(config *sdk.Config) (*unet.UNet, *umon.UMon) {
	var publicKey, privateKey string
	for _, env := range os.Environ() {
		pair := strings.Split(env, "=")
		if pair[0] == "PUBLIC_KEY" {
			publicKey = pair[1]
			continue
		}
		if pair[0] == "PRIVATE_KEY" {
			privateKey = pair[1]
			continue
		}
	}

	if publicKey == "" {
		log.Println("Loading PUBLIC_KEY from config file")
		publicKey = config.Global.PublicKey
	}

	if privateKey == "" {
		log.Println("Loading PRIVATE_KEY from config file")
		privateKey = config.Global.PrivateKey
	}

	uNet = unet.New(&ucloud.Config{
		Credentials: &auth.KeyPair{
			PublicKey:  publicKey,
			PrivateKey: privateKey,
		},
	})
	uMon = umon.New(&ucloud.Config{
		Credentials: &auth.KeyPair{
			PublicKey:  publicKey,
			PrivateKey: privateKey,
		},
	})
	return uNet, uMon
}
