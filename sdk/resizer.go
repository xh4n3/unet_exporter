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

func (r *Resizer) SetToAdvisedBandwidth() error {
	advisedBandwidth := r.AdvisedBandwidth()
	if r.verbose {
		log.Printf("Advisor suggests %v", advisedBandwidth)
	}
	upLimit, downLimit := r.CurrentLimit()

	if advisedBandwidth <= upLimit && advisedBandwidth >= downLimit {
		return r.SetCurrentBandwidth(advisedBandwidth)
	} else {
		log.Println("Advised bandwidth exceeded limit.")
		if advisedBandwidth <= downLimit {
			return r.SetCurrentBandwidth(downLimit)
		}
		if advisedBandwidth >= upLimit {
			return r.SetCurrentBandwidth(upLimit)
		}
	}
}

func (r *Resizer) CurrentLimit() (int, int) {
	log.Printf("Hardlimit Up: %v	Down: %v", r.target.HardLimit.UpLimit, r.target.HardLimit.DownLimit)
	return r.target.HardLimit.UpLimit, r.target.HardLimit.DownLimit
}

func (r *Resizer) AdvisedBandwidth() int {
	bandwidthLimits := []int{}

	for _, query := range r.target.VariedLimits {
		bandwidthLimit, err := r.RunQuery(query)
		if err != nil {
			log.Println(err)
		} else {
			bandwidthLimits = append(bandwidthLimits, bandwidthLimit)
		}
	}

	// highestLimit might be zero, so check if it is between downLimit and upLimit
	if highestLimit := highest(bandwidthLimits); highestLimit >= 0 {
		return int((100 + r.target.RaiseRatio) * highestLimit / 100)
	} else {
		return r.target.DefaultBandwidth
	}
}

func (r *Resizer) RunQuery(query string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
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

func highest(numbers []int) int {
	var highest int
	for _, number := range numbers {
		if number > highest {
			highest = number
		}
	}
	return highest
}
