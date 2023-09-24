package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/numaproj/numaflow-go/pkg/sourcer"

	"numaflow-nats-source/pkg/config"
	"numaflow-nats-source/pkg/nats"
	"numaflow-nats-source/pkg/utils"
)

func main() {
	// Get the config file path and format from env vars
	var format string
	format, ok := os.LookupEnv("CONFIG_FORMAT")
	if !ok {
		log.Printf("CONFIG_FORMAT not set, defaulting to yaml")
		format = "yaml"
	}

	var config *config.Config
	var err error

	config, err = getConfigFromEnvVars(format)
	if err != nil {
		log.Panic("Failed to parse config from env vars : ", err)
		config, err = getConfigFromFile(format)
		if err != nil {
			log.Panic("Failed to parse config file : ", err)
		} else {
			log.Printf("Successfully parsed config file")
		}
	} else {
		log.Printf("Successfully parsed config from env vars")
	}

	natsSrc, err := nats.New(config)
	if err != nil {
		log.Panic("Failed to create nats source : ", err)
	}
	defer natsSrc.Close()

	err = sourcer.NewServer(natsSrc).Start(context.Background())
	if err != nil {
		log.Panic("Failed to start source server : ", err)
	}
}

func getConfigFromFile(format string) (*config.Config, error) {
	if format == "yaml" {
		parser := &config.YAMLConfigParser{}
		content, err := os.ReadFile(fmt.Sprintf("%s/nats-config.yaml", utils.ConfigVolumePath))
		if err != nil {
			return nil, err
		}
		return parser.Parse(string(content))
	} else if format == "json" {
		parser := &config.JSONConfigParser{}
		content, err := os.ReadFile(fmt.Sprintf("%s/nats-config.json", utils.ConfigVolumePath))
		if err != nil {
			return nil, err
		}
		return parser.Parse(string(content))
	} else {
		return nil, fmt.Errorf("invalid config format %s", format)
	}
}

func getConfigFromEnvVars(format string) (*config.Config, error) {
	var c string
	c, ok := os.LookupEnv("NATS_CONFIG")
	if !ok {
		return nil, fmt.Errorf("NATS_CONFIG environment variable is not set")
	}
	if format == "yaml" {
		parser := &config.YAMLConfigParser{}
		return parser.Parse(c)
	} else if format == "json" {
		parser := &config.JSONConfigParser{}
		return parser.Parse(c)
	} else {
		return nil, fmt.Errorf("invalid config format %s", format)
	}
}
