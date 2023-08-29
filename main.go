package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/numaproj/numaflow-go/pkg/sourcer"

	"numaflow-nats-source/impl"
	"numaflow-nats-source/pkg/configuration"
)

func main() {
	configFilePath := "/etc/config/nats-config.json"
	if err := printFileContent(configFilePath); err != nil {
		log.Fatalf("Failed to print config file %s : %v", configFilePath, err)
	}

	config, err := getConfigFromFile(configFilePath)
	if err != nil {
		log.Fatalf("Failed to parse config file %s : %v", configFilePath, err)
	} else {
		log.Printf("Parsed config file %s : %v", configFilePath, config)
	}

	var prefixStr string
	flag.StringVar(&prefixStr, "prefix", "test-payload-", "Prefix of the payload")
	flag.Parse()
	simpleSource := impl.NewSimpleSource(prefixStr)
	err = sourcer.NewServer(simpleSource).Start(context.Background())
	if err != nil {
		log.Panic("Failed to start source server : ", err)
	}
}

func printFileContent(filePath string) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	fmt.Print(string(content))
	return nil
}

func getConfigFromFile(filePath string) (*configuration.Config, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	c := &configuration.Config{}
	if err = c.Parse(string(content)); err != nil {
		return nil, err
	} else {
		return c, nil
	}
}
