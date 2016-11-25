package sdk

import (
	"fmt"
	"github.com/xh4n3/ucloud-sdk-go/service/unet"
	"log"
)

type Collector struct {
	target         *Target
	uNet           *unet.UNet
	eipResourceMap map[string]string
}

func NewCollector(uNet *unet.UNet, target *Target) *Collector {
	return &Collector{
		target: target,
		uNet:   uNet,
	}
}

func (c *Collector) GetCurrentBandwidth() int {
	shareBandwidthResp, err := c.uNet.DescribeShareBandwidth(&unet.DescribeShareBandwidthParams{
		Region: c.target.Region,
	})

	if err != nil {
		log.Fatal(err)
	}

	for _, shareBandwidth := range *shareBandwidthResp.DataSet {
		if shareBandwidth.ShareBandwidthId == c.target.Name {
			return shareBandwidth.ShareBandwidth
		}
	}
	return -1
}

func (c *Collector) ListEIPs() {
	eipsResp, err := c.uNet.DescribeEIP(&unet.DescribeEIPParams{
		Region: c.target.Region,
	})
	if err != nil {
		log.Fatal(err)
	}

	c.eipResourceMap = make(map[string]string)

	for _, eip := range *(eipsResp.EIPSet) {
		c.eipResourceMap[eip.EIPId] = bandwidthLabel(eip)
	}
}

func (c *Collector) ListBandwidthUsages() (map[string]float32, float32){
	usageResp, err := c.uNet.DescribeBandwidthUsage(&unet.DescribeBandwidthUsageParams{
		Region: c.target.Region,
	})

	if err != nil {
		log.Fatal(err)
	}
	resourceBandwidthMap := make(map[string]float32)

	for _, bandwidth := range *usageResp.EIPSet {
		if resourceName, ok := c.eipResourceMap[bandwidth.EIPId]; ok {
			resourceBandwidthMap[resourceName] = bandwidth.CurBandwidth
		} else {
			log.Printf("cannot find resourceName for EIP %v\n", bandwidth.EIPId)
		}
	}

	bandwidthTotalUsed := float32(0)
	for _, eip := range *usageResp.EIPSet {
		bandwidthTotalUsed += eip.CurBandwidth
	}
	return resourceBandwidthMap, bandwidthTotalUsed
}


func bandwidthLabel(eipset unet.EIPSet) string {
	ips := ""
	for _, ip := range *eipset.EIPAddr {
		ips += "_" + ip.IP
	}
	return fmt.Sprintf("%v%v", eipset.Resource.ResourceType, ips)

}