package sdk

import (
	"context"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/api/prometheus"
	"github.com/prometheus/common/model"
	"github.com/xh4n3/ucloud-sdk-go/service/unet"
	"log"
	"strconv"
	"time"
)

type Resizer struct {
	target           *Target
	apiClient        prometheus.QueryAPI
	uNet             *unet.UNet
	verbose          bool
	dryRun           bool
	lastSetBandwidth int
}

func NewResizer(uNet *unet.UNet, target *Target, config *Config) *Resizer {
	client, err := prometheus.New(prometheus.Config{
		Address: target.QueryEndpoint,
	})
	if err != nil {
		log.Fatal(err)
	}
	apiClient := prometheus.NewQueryAPI(client)
	return &Resizer{
		target:    target,
		apiClient: apiClient,
		verbose:   config.Global.Verbose,
		uNet:      uNet,
		dryRun:    config.Global.DryRun,
	}
}

func (r *Resizer) SetCurrentBandwidth(newBandwidth int) error {
	if r.lastSetBandwidth == newBandwidth {
		return nil
	}

	log.Printf("Switching %v's bandwidth to %v\n", r.target.Name, newBandwidth)
	if r.dryRun {
		log.Println("dryRun enabled, nothing will be done.")
		return nil
	}

	_, err := r.uNet.ResizeShareBandwidth(&unet.ResizeShareBandwidthParams{
		Region:           r.target.Region,
		ShareBandwidth:   newBandwidth,
		ShareBandwidthId: r.target.Name,
	})
	if err != nil {
		return err
	}

	r.lastSetBandwidth = newBandwidth

	return nil
}

func (r *Resizer) SetToAdvisedBandwidth() {
	advisedBandwidth := r.AdvisedBandwidth()
	if r.verbose {
		log.Printf("Advisor suggests %v", advisedBandwidth)
	}
	upLimit, downLimit := r.CurrentLimit()
	if r.verbose {
		log.Printf("bandwidth limit now %v %v", upLimit, downLimit)
	}
	if advisedBandwidth <= upLimit && advisedBandwidth >= downLimit {
		r.SetCurrentBandwidth(advisedBandwidth)
	} else {
		log.Println("Advised bandwidth exceeded limit.")
	}
}

func (r *Resizer) CurrentLimit() (int, int) {
	now := time.Now()
	hourNow, _, _ := now.Clock()
	weekdayNow := int(now.Weekday())

	var defaultLimit *VariedLimit

	for _, limit := range r.target.VariedLimits {
		if limit.Name == "default" {
			defaultLimit = limit
		}
		if contains(weekdayNow, limit.WeekDays) && contains(hourNow, limit.Hours) {
			if r.verbose {
				log.Printf("Limit template: %v	UpLimit: %v	DownLimit: %v", limit.Name, limit.UpLimit, limit.DownLimit)
			}
			return limit.UpLimit, limit.DownLimit
		}
	}
	if defaultLimit == nil {
		log.Fatalln("No default limit specified")
	}

	log.Printf("Limit template: default	UpLimit: %v	DownLimit: %v", defaultLimit.UpLimit, defaultLimit.DownLimit)
	return defaultLimit.UpLimit, defaultLimit.DownLimit
}

func (r *Resizer) AdvisedBandwidth() int {
	bandwidthLimits := []int{}

	oneWeekAgo := "max_over_time(total_bandwidth_usage[30m] offset 7d)"
	oneWeekAgoBandwidth, err := r.RunQuery(oneWeekAgo)
	if err != nil {
		log.Println(err)
	} else {
		bandwidthLimits = append(bandwidthLimits, oneWeekAgoBandwidth)
	}

	oneDayAgo := "max_over_time(total_bandwidth_usage[30m] offset 1d)"
	oneDayAgoBandwidth, err := r.RunQuery(oneDayAgo)
	if err != nil {
		log.Println(err)
	} else {
		bandwidthLimits = append(bandwidthLimits, oneDayAgoBandwidth)
	}

	current := "max_over_time(total_bandwidth_usage[10m])"
	currentBandwidth, err := r.RunQuery(current)
	if err != nil {
		log.Println(err)
	} else {
		bandwidthLimits = append(bandwidthLimits, currentBandwidth)
	}

	if highestLimit := highest(bandwidthLimits); highestLimit > 0 {
		return int((100 + r.target.RaiseRatio) * highestLimit / 100)
	} else {
		return r.target.DefaultBandwidth
	}
}

func (r *Resizer) RunQuery(query string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10 * time.Second))
	defer cancel()
	val, err := r.apiClient.Query(ctx, query, time.Now())
	if err != nil {
		return 0, err
	}
	if vector, ok := val.(model.Vector); ok && vector.Len() > 0 {
		for _, sample := range val.(model.Vector) {
			downLimitFloat, err := strconv.ParseFloat(sample.Value.String(), 32)
			if err != nil {
				return 0, err
			}
			downLimit := int(downLimitFloat)
			return downLimit, nil
		}
	}
	return 0, errors.New("query failed")
}

func contains(number int, numbers []int) bool {
	for _, num := range numbers {
		if num == number {
			return true
		}
	}
	return false
}

func highest(numbers []int) int {
	var highest int
	for _, number := range numbers {
		if number > highest {
			highest = number
		}
	}
	return highest
}
