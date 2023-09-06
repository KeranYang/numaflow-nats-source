package utils

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

const (
	TestSecretVolumePath = "./test-secret-path"
)

func Test_GetSecretVolumePath(t *testing.T) {
	underTest := NewNatsVolumeReader(TestSecretVolumePath)
	testSecretKeySelector := &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: "test-secret",
		},
		Key: "test-key",
	}
	p, e := underTest.GetSecretVolumePath(testSecretKeySelector)
	assert.Nil(t, e)
	assert.Equal(t, fmt.Sprintf("%s/test-secret/test-key", TestSecretVolumePath), p)
}

func Test_GetSecretFromVolume(t *testing.T) {
	underTest := NewNatsVolumeReader(TestSecretVolumePath)
	testSecretKeySelector := &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: "test-secret",
		},
		Key: "test-key",
	}

	// Prepare
	data := []byte("test-data")
	err := os.MkdirAll(fmt.Sprintf("%s/test-secret", TestSecretVolumePath), 0750)
	assert.Nil(t, err)
	file, err := os.Create(fmt.Sprintf("%s/test-secret/test-key", TestSecretVolumePath))
	assert.Nil(t, err)
	defer file.Close()
	_, err = file.Write(data)
	assert.Nil(t, err)

	// Test
	v, err := underTest.GetSecretFromVolume(testSecretKeySelector)
	assert.Nil(t, err)

	// Verify
	assert.Equal(t, string(data), v)

	// Cleanup
	err = os.RemoveAll(TestSecretVolumePath)
	assert.NoError(t, err)
}
