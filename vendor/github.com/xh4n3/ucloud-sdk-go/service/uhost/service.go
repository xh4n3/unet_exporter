package uhost

import (
	"net/http"

	"github.com/xh4n3/ucloud-sdk-go/ucloud"
	"github.com/xh4n3/ucloud-sdk-go/ucloud/service"
)

type UHost struct {
	*service.Service
}

func New(config *ucloud.Config) *UHost {

	service := &service.Service{
		Config:      ucloud.DefaultConfig.Merge(config),
		ServiceName: "UHost",
		APIVersion:  ucloud.APIVersion,

		BaseUrl:    ucloud.APIBaseURL,
		HttpClient: &http.Client{},
	}

	return &UHost{service}

}
