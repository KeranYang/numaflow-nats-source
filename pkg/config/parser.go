package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/yaml.v2"
)

// Parser is an interface that defines methods to parse and un-parse Config objects.
type Parser interface {
	Parse(configString string) (*NatsConfig, error)
	UnParse(config *NatsConfig) (string, error)
}

// YAMLConfigParser is a parser for YAML formatted configuration strings
type YAMLConfigParser struct{}

func (p *YAMLConfigParser) Parse(configString string) (*NatsConfig, error) {
	c := &NatsConfig{}
	err := yaml.Unmarshal([]byte(configString), c)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config string: %w", err)
	}
	return c, nil
}

func (p *YAMLConfigParser) UnParse(config *NatsConfig) (string, error) {
	if config == nil {
		return "", errors.New("config cannot be nil")
	}
	b, err := yaml.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to un-parse config: %w", err)
	}
	return string(b), nil
}

// JSONConfigParser is a parser for JSON formatted configuration strings.
type JSONConfigParser struct{}

func (p *JSONConfigParser) Parse(configString string) (*NatsConfig, error) {
	c := &NatsConfig{}
	err := json.Unmarshal([]byte(configString), c)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config string: %w", err)
	}
	return c, nil
}

func (p *JSONConfigParser) UnParse(config *NatsConfig) (string, error) {
	if config == nil {
		return "", errors.New("config cannot be nil")
	}
	b, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to un-parse config: %w", err)
	}
	return string(b), nil
}
