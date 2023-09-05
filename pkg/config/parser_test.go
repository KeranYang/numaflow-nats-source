package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestConfigParser(t *testing.T) {
	var parsers = []Parser{
		&JSONConfigParser{},
		&YAMLConfigParser{},
	}
	for _, parser := range parsers {
		testConfig := &Config{
			URL:     "nats://localhost:4222",
			Subject: "test",
			Queue:   "test",
			TLS: &TLS{
				InsecureSkipVerify: true,
			},
			Auth: &Auth{
				Basic: &BasicAuth{
					User: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "test",
						},
						Key: "test",
					},
				},
			},
		}
		configStr, err := parser.UnParse(testConfig)
		assert.NoError(t, err)
		config, err := parser.Parse(configStr)
		assert.NoError(t, err)
		assert.Equal(t, testConfig, config)
	}
}
