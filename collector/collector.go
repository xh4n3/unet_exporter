package collector

import (
	"fmt"
	"github.com/xh4n3/ucloud-sdk-go/service/unet"
	"github.com/xh4n3/ucloud-sdk-go/ucloud/utils"
	"log"
)

type Collector struct {
	target *Target
	uNet   *unet.UNet
}

func NewCollector(uNet *unet.UNet, target *Target) *Collector {
	return &Collector{
		target: target,
		uNet:   uNet,
	}
}

func (c *Collector) ListShareBandwidth() {
	shareBandwidthResp, err := c.uNet.DescribeShareBandwidth(&unet.DescribeShareBandwidthParams{
		Region: c.target.Region,
	})

	if err != nil {
		log.Fatal(err)
	}

	utils.DumpVal(shareBandwidthResp)
}

func (c *Collector) ListEIPs() {
	eipsResp, err := c.uNet.DescribeEIP(&unet.DescribeEIPParams{
		Region: c.target.Region,
	})
	if err != nil {
		log.Fatal(err)
	}

	eipResourceMap := make(map[string]string)

	for _, eip := range *(eipsResp.EIPSet) {
		eipResourceMap[eip.EIPId] = fmt.Sprintf("%v_%v_%v", eip.Resource.Zone, eip.Resource.ResourceType, eip.Resource.ResourceName)
	}

}

func (c *Collector) ListBandwidthUsages() {

	usageResp, err := c.uNet.DescribeBandwidthUsage(&unet.DescribeBandwidthUsageParams{
		Region: c.target.Region,
	})

	if err != nil {
		log.Fatal(err)
	}
	utils.DumpVal(usageResp)
	bandwidthUsed := float32(0)
	for _, eip := range *usageResp.EIPSet {
		bandwidthUsed += eip.CurBandwidth
	}
	log.Println(bandwidthUsed)
}
