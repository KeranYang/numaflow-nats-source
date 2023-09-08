package config

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestConfigParser_UnParseThenParse(t *testing.T) {
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

func TestConfigParser_ParseErrScenarios(t *testing.T) {
	var parsers = []Parser{
		&JSONConfigParser{},
		&YAMLConfigParser{},
	}
	for _, parser := range parsers {
		_, err := parser.Parse("invalid config string")
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "failed to parse config string"))
		_, err = parser.UnParse(nil)
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "config cannot be nil"))
	}
}

func TestConfigParser_YAML(t *testing.T) {
	yamlStr := `
url: nats
subject: test-subject
queue: my-queue
auth:
  token:
    localobjectreference:
      name: nats-auth-fake-token
    key: fake-token
`
	parser := &YAMLConfigParser{}
	config, err := parser.Parse(yamlStr)
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(&Config{
		URL:     "nats",
		Subject: "test-subject",
		Queue:   "my-queue",
		Auth: &Auth{
			Token: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "nats-auth-fake-token",
				},
				Key: "fake-token",
			},
		},
	}, config))
}

func TestConfigParser_JSON(t *testing.T) {
	jsonStr := `
{
  "url":"nats",
  "subject":"test-subject",
  "queue":"my-queue",
  "auth":{
    "token":{
      "name":"nats-auth-fake-token",
      "key":"fake-token"
    }
  }
}
`
	parser := &JSONConfigParser{}
	config, err := parser.Parse(jsonStr)
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(&Config{
		URL:     "nats",
		Subject: "test-subject",
		Queue:   "my-queue",
		Auth: &Auth{
			Token: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "nats-auth-fake-token",
				},
				Key: "fake-token",
			},
		},
	}, config))
}
