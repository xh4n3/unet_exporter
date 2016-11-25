package sdk

import (
	"github.com/pkg/errors"
	"github.com/xh4n3/ucloud-sdk-go/service/unet"
	"log"
)

type Resizer struct {
	target           *Target
	uNet             *unet.UNet
	currentBandwidth int
}

func NewResizer(uNet *unet.UNet, target *Target) *Resizer {
	return &Resizer{
		target: target,
		uNet:   uNet,
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
	log.Println(newBandwidth)
	return nil
}

func (r *Resizer) IncreaseBandwidth() error {
	currentBandwidth, err := r.GetCurrentBandwidth()
	if err != nil {
		log.Printf("unable to get current bandwidth for shareBandwidth %v: %v", r.target.Name, err)
	}
	if currentBandwidth + r.target.Step <= r.target.UpLimit {
		r.SetCurrentBandwidth(currentBandwidth + r.target.Step)
	} else {
		return errors.New("uplimit hit")
	}
	return nil
}

func (r *Resizer) DecreaseBandwidth() error {
	currentBandwidth, err := r.GetCurrentBandwidth()
	if err != nil {
		log.Printf("unable to get current bandwidth for shareBandwidth %v: %v", r.target.Name, err)
	}
	if currentBandwidth - r.target.Step >= r.target.DownLimit {
		r.SetCurrentBandwidth(currentBandwidth - r.target.Step)
	} else {
		return errors.New("downlimit hit")
	}
	return nil
}
