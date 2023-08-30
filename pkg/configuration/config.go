package configuration

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
)

type Auth struct {
	// Token auth
	// +optional
	Token *corev1.SecretKeySelector `json:"token,omitempty" protobuf:"bytes,1,opt,name=token"`
	// TODO - NKey auth
	// TODO - Basic auth which contains a user name and a password
}

// TODO - TLS

type Config struct {
	// URL to connect to NATS cluster, multiple urls could be separated by comma.
	URL string `json:"url" protobuf:"bytes,1,opt,name=url"`
	// Subject holds the name of the subject onto which messages are published.
	Subject string `json:"subject" protobuf:"bytes,2,opt,name=subject"`
	// Queue is used for queue subscription.
	Queue string `json:"queue" protobuf:"bytes,3,opt,name=queue"`
	// Auth information
	// +optional
	Auth *Auth `json:"auth,omitempty" protobuf:"bytes,4,opt,name=auth"`
}

// String returns the string representation of the Config.
func (c *Config) String() string {
	b, err := json.Marshal(c)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func (c *Config) Parse(s string) error {
	return json.Unmarshal([]byte(s), c)
}
