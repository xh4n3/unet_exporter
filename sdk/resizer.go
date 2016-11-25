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
	log.Printf("Switching bandwidth to %v\n", newBandwidth)
	_, err := r.uNet.ResizeShareBandwidth(&unet.ResizeShareBandwidthParams{
		Region: r.target.Region,
		ShareBandwidth: newBandwidth,
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
	if currentBandwidth + r.target.Step <= r.target.UpLimit {
		return r.SetCurrentBandwidth(currentBandwidth + r.target.Step)
	} else {
		return errors.New("uplimit hit")
	}
}

func (r *Resizer) DecreaseBandwidth() error {
	currentBandwidth, err := r.GetCurrentBandwidth()
	if err != nil {
		return errors.Errorf("unable to get current bandwidth for shareBandwidth %v: %v", r.target.Name, err)
	}
	if currentBandwidth - r.target.Step >= r.target.DownLimit {
		return r.SetCurrentBandwidth(currentBandwidth - r.target.Step)
	} else {
		return errors.New("downlimit hit")
	}
}
