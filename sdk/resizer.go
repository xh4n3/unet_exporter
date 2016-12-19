package sdk

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/xh4n3/ucloud-sdk-go/service/unet"
	"log"
	"time"
	"os/exec"
	"strconv"
	"strings"
)

type Resizer struct {
	target           *Target
	uNet             *unet.UNet
	dryRun           bool
	downLimitAdvisor string
	currentBandwidth int
}

func NewResizer(uNet *unet.UNet, target *Target, config *Config) *Resizer {
	return &Resizer{
		target: target,
		uNet:   uNet,
		dryRun: config.Global.DryRun,
		downLimitAdvisor: config.Plugins.DownLimitAdvisor,
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
			log.Printf("Limit template: %v	UpLimit: %v	DownLimit: %v	UpStep: %v	DownStep: %v", limit.Name, limit.UpLimit, higher(limit.DownLimit, r.AdvisedDownLimit()), limit.UpStep, limit.DownStep)
			return limit.UpLimit, higher(limit.DownLimit, r.AdvisedDownLimit()), limit.UpStep, limit.DownStep
		}
	}
	if defaultLimit == nil {
		log.Fatalln("No default limit specified")
	}

	log.Printf("Limit template: default	UpLimit: %v	DownLimit: %v	UpStep: %v	DownStep: %v", defaultLimit.UpLimit, higher(defaultLimit.DownLimit, r.AdvisedDownLimit()), defaultLimit.UpStep, defaultLimit.DownStep)
	return defaultLimit.UpLimit, higher(defaultLimit.DownLimit, r.AdvisedDownLimit()), defaultLimit.UpStep, defaultLimit.DownStep
}

func (r *Resizer) AdvisedDownLimit() int {
	downLimit := 0
	if r.downLimitAdvisor != "" {
		output, err := exec.Command("/bin/sh", r.downLimitAdvisor).Output()
		if err != nil {
			log.Printf("Limit Advisor Failed: %v", err.Error())
			return 0
		}
		downLimit, err = strconv.Atoi(strings.TrimSpace(string(output)))
		if err != nil {
			log.Printf("Read Limit Advisor Result Failed: %v", err.Error())
			return 0
		}
		log.Printf("Limit Advisor suggests down limit at %v", downLimit)
	}
	return downLimit
}

func contains(number int, numbers []int) bool {
	for _, num := range numbers {
		if num == number {
			return true
		}
	}
	return false
}

func higher(number1, number2 int) int {
	if number1 > number2 {
		return number1
	}
	return number2
}
