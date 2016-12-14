package sdk

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/xh4n3/ucloud-sdk-go/service/unet"
	"log"
	"time"
)

type Resizer struct {
	target           *Target
	uNet             *unet.UNet
	dryRun           bool
	currentBandwidth int
}

func NewResizer(uNet *unet.UNet, target *Target, dryRun bool) *Resizer {
	return &Resizer{
		target: target,
		uNet:   uNet,
		dryRun: dryRun,
	}
}

func (r *Resizer) GetCurrentBandwidth() (int, error) {
	shareBandwidthResp, err := r.uNet.DescribeShareBandwidth(&unet.DescribeShareBandwidthParams{
		Region: r.target.Region,
	})

	if err != nil {
		log.Fatal(err)
	}

	for _, shareBandwidth := range *shareBandwidthResp.DataSet {
		if shareBandwidth.ShareBandwidthId == r.target.Name {
			return shareBandwidth.ShareBandwidth, nil
		}
	}
	return 0, errors.New("cannot find shareBandwidth")
}

func (r *Resizer) SetCurrentBandwidth(newBandwidth int) error {
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

	return nil
}

func (r *Resizer) IncreaseBandwidth() error {
	currentBandwidth, err := r.GetCurrentBandwidth()
	if err != nil {
		return errors.Errorf("unable to get current bandwidth for shareBandwidth %v: %v", r.target.Name, err)
	}
	upLimit, _, upStep, _ := r.CurrentLimitAndStep()

	if currentBandwidth+upStep <= upLimit {
		return r.SetCurrentBandwidth(currentBandwidth + upStep)
	} else {
		return fmt.Errorf("uplimit hit at %v", upLimit)
	}
}

func (r *Resizer) DecreaseBandwidth() error {
	currentBandwidth, err := r.GetCurrentBandwidth()
	if err != nil {
		return errors.Errorf("unable to get current bandwidth for shareBandwidth %v: %v", r.target.Name, err)
	}
	_, downLimit, _, downStep := r.CurrentLimitAndStep()

	if currentBandwidth-downStep >= downLimit {
		return r.SetCurrentBandwidth(currentBandwidth - downStep)
	} else {
		return fmt.Errorf("downlimit hit at %v", downLimit)
	}
}

func (r *Resizer) CurrentLimitAndStep() (int, int, int, int) {
	now := time.Now()
	hourNow, _, _ := now.Clock()
	weekdayNow := int(now.Weekday())

	var defaultLimit *VariedLimit

	for _, limit := range r.target.VariedLimits {
		if limit.Name == "default" {
			defaultLimit = limit
		}
		if contains(weekdayNow, limit.WeekDays) && contains(hourNow, limit.Hours) {
			log.Printf("Limit template: %v	UpLimit: %v	DownLimit: %v	UpStep: %v	DownStep: %v", limit.Name, limit.UpLimit, limit.DownLimit, limit.UpStep, limit.DownStep)
			return limit.UpLimit, limit.DownLimit, limit.UpStep, limit.DownStep
		}
	}
	if defaultLimit == nil {
		log.Fatalln("No default limit specified")
	}

	log.Printf("Limit template: default	UpLimit: %v	DownLimit: %v	UpStep: %v	DownStep: %v", defaultLimit.UpLimit, defaultLimit.DownLimit, defaultLimit.UpStep, defaultLimit.DownStep)
	return defaultLimit.UpLimit, defaultLimit.DownLimit, defaultLimit.UpStep, defaultLimit.DownStep
}

func contains(number int, numbers []int) bool {
	for _, num := range numbers {
		if num == number {
			return true
		}
	}
	return false
}
