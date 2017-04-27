package umon

import (
	"github.com/pkg/errors"
	"github.com/xh4n3/ucloud-sdk-go/ucloud"
)

type GetMetricParams struct {
	ucloud.CommonRequest

	Region       string
	Zone         string
	ResourceType string
	ResourceId   string
	TimeRange    int
	BeginTime    int64
	EndTime      int64
	Statistic    string
	MetricName   []string
}

type GetMetricResponseOfShareBandwidth struct {
	ucloud.CommonResponse

	DataSets *GetMetricShareBandwidthDataSet
}

type GetMetricShareBandwidthDataSet struct {
	BandIn  *[]GetMetricShareBandwidthDataItem
	BandOut *[]GetMetricShareBandwidthDataItem
}

type GetMetricShareBandwidthDataItem struct {
	Value     int64
	Timestamp int64
}

func (u *UMon) GetMetric(params *GetMetricParams) (*GetMetricResponseOfShareBandwidth, error) {
	if params.ResourceType == "sharebandwidth" {
		response := &GetMetricResponseOfShareBandwidth{}
		err := u.DoRequest("GetMetric", params, response)
		return response, err
	} else {
		return nil, errors.New("ResourceType error")
	}
}
