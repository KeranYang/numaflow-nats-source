package utils

import (
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// GetSecretFromVolume retrieves the value of mounted secret volume
// "/var/numaflow/secrets/${secretRef.name}/${secretRef.key}" is expected to be the file path
func GetSecretFromVolume(selector *corev1.SecretKeySelector) (string, error) {
	filePath, err := GetSecretVolumePath(selector)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get secret value of name: %s, key: %s, %w", selector.Name, selector.Key, err)
	}
	// Secrets edited by tools like "vim" always have an extra invisible "\n" in the end,
	// and it's often neglected, but it makes differences for some of the applications.
	return strings.TrimSuffix(string(data), "\n"), nil
}

// GetSecretVolumePath returns the path of the mounted secret
func GetSecretVolumePath(selector *corev1.SecretKeySelector) (string, error) {
	if selector == nil {
		return "", fmt.Errorf("secret key selector is nil")
	}
	return fmt.Sprintf("/etc/secrets/%s/%s", selector.Name, selector.Key), nil
}
