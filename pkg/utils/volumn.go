package utils

import (
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

type VolumeReader interface {
	GetSecretFromVolume(selector *corev1.SecretKeySelector) (string, error)
	GetSecretVolumePath(selector *corev1.SecretKeySelector) (string, error)
}

type NatsVolumeReader struct {
	secretPath string
}

func NewNatsVolumeReader(secretPath string) *NatsVolumeReader {
	return &NatsVolumeReader{
		secretPath: secretPath,
	}
}

// GetSecretFromVolume retrieves the value of mounted secret volume
func (nvr *NatsVolumeReader) GetSecretFromVolume(selector *corev1.SecretKeySelector) (string, error) {
	filePath, err := nvr.GetSecretVolumePath(selector)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get secret value of name: %s, key: %s, %w", selector.Name, selector.Key, err)
	}
	// Secrets edited by tools like "vim" always have an extra invisible "\n" in the end,
	// and it's often neglected, but it makes differences for some applications.
	return strings.TrimSuffix(string(data), "\n"), nil
}

// GetSecretVolumePath returns the path of the mounted secret
func (nvr *NatsVolumeReader) GetSecretVolumePath(selector *corev1.SecretKeySelector) (string, error) {
	if selector == nil {
		return "", fmt.Errorf("secret key selector is nil")
	}
	return fmt.Sprintf("%s/%s/%s", nvr.secretPath, selector.Name, selector.Key), nil
}
