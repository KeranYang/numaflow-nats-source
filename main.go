package main

import (
	"context"
	"io/ioutil"
	"log"

	"github.com/numaproj/numaflow-go/pkg/sourcer"

	"numaflow-nats-source/pkg/config"
	"numaflow-nats-source/pkg/nats"
)

func main() {
	configFilePath := "/etc/config/nats-config.json"
	config, err := getConfigFromFile(configFilePath)
	if err != nil {
		log.Fatalf("Failed to parse config file %s : %v", configFilePath, err)
	} else {
		log.Printf("Parsed config file %s : %v", configFilePath, config)
	}
	natsSrc, err := nats.New(config)
	if err != nil {
		log.Panic("Failed to create nats source : ", err)
	}
	err = sourcer.NewServer(natsSrc).Start(context.Background())
	if err != nil {
		log.Panic("Failed to start source server : ", err)
	}
}

func getConfigFromFile(filePath string) (*config.Config, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	c := &config.Config{}
	if err = c.Parse(string(content)); err != nil {
		return nil, err
	} else {
		return c, nil
	}
}
