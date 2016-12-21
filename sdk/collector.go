package sdk

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/xh4n3/ucloud-sdk-go/service/umon"
	"github.com/xh4n3/ucloud-sdk-go/service/unet"
	"log"
	"strings"
)

type Collector struct {
	target         *Target
	uNet           *unet.UNet
	uMon           *umon.UMon
	eipResourceMap map[string]string
}

func NewCollector(uNet *unet.UNet, uMon *umon.UMon, target *Target) *Collector {
	return &Collector{
		target: target,
		uNet:   uNet,
		uMon:   uMon,
	}
}

func (c *Collector) GetCurrentBandwidth() (int, error) {
	shareBandwidthResp, err := c.uNet.DescribeShareBandwidth(&unet.DescribeShareBandwidthParams{
		Region: c.target.Region,
	})

	if err != nil {
		return 0, err
	}

	for _, shareBandwidth := range *shareBandwidthResp.DataSet {
		if shareBandwidth.ShareBandwidthId == c.target.Name {
			return shareBandwidth.ShareBandwidth, nil
		}
	}
	return -1, errors.New("cannot find target shareBandwidth")
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

func (c *Collector) ListBandwidthUsages() map[string]float32 {
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
			continue
		}
	}

	return resourceBandwidthMap
}

func (c *Collector) GetTotalBandwidth() float64 {
	metricResponse, err := c.uMon.GetMetric(&umon.GetMetricParams{
		Region:       c.target.Region,
		ResourceType: "sharebandwidth",
		ResourceId:   c.target.Name,
		MetricName:   []string{"BandIn", "BandOut"},
		TimeRange:    120,
	})
	if err != nil {
		log.Fatal(err)
	}
	bandIn := *(metricResponse.DataSets.BandIn)
	bandOut := *(metricResponse.DataSets.BandOut)
	if len(bandIn) > 0 && len(bandOut) > 0 {
		total := bandIn[len(bandIn)-1].Value + bandOut[len(bandOut)-1].Value
		return float64(total / 1024 / 1024)
	}
	return float64(0)
}

func bandwidthLabel(eipset unet.EIPSet) string {
	ips := ""
	for _, ip := range *eipset.EIPAddr {
		ips += "_" + ip.IP
	}
	return fmt.Sprintf("%v%v", eipset.Resource.ResourceType, ips)

}
