package main

import (
	"github.com/xh4n3/ucloud-sdk-go/service/unet"
	"github.com/xh4n3/ucloud-sdk-go/ucloud"
	"github.com/xh4n3/ucloud-sdk-go/ucloud/auth"
	"github.com/xh4n3/unet_exporter/collector"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

func main() {
	configContent, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Fatalf("Config file not found: %v", err)
	}

	config := &collector.Config{}
	err = yaml.Unmarshal(configContent, config)
	if err != nil {
		log.Fatalf("cannot unmarshal config file: %v", err)
	}

	uNet := unet.New(&ucloud.Config{
		Credentials: &auth.KeyPair{
			PublicKey:  config.Global.PublicKey,
			PrivateKey: config.Global.PrivateKey,
		},
	})

	collector := collector.NewCollector(uNet, config.Targets[0])
	collector.ListShareBandwidth()

}
