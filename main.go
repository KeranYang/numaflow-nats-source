package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/numaproj/numaflow-go/pkg/sourcer"
)

func main() {
	configFilePath := "/etc/config/nats-config.json"
	if err := printFileContent(configFilePath); err != nil {
		log.Fatalf("Failed to print config file %s : %v", configFilePath, err)
	}

	var prefixStr string
	flag.StringVar(&prefixStr, "prefix", "test-payload-", "Prefix of the payload")
	flag.Parse()
	simpleSource := NewSimpleSource(prefixStr)
	err := sourcer.NewServer(simpleSource).Start(context.Background())
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
