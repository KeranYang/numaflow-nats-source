package nats

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	natstestserver "github.com/nats-io/nats-server/v2/test"
	natslib "github.com/nats-io/nats.go"
	sourcesdk "github.com/numaproj/numaflow-go/pkg/sourcer"
	"github.com/stretchr/testify/assert"

	"numaflow-nats-source/pkg/config"
)

type TestReadRequest struct {
	count   uint64
	timeout time.Duration
}

func (rr TestReadRequest) Count() uint64 {
	return rr.count
}

func (rr TestReadRequest) TimeOut() time.Duration {
	return rr.timeout
}

// Test_Single tests a single source reading from a single nats subject
func Test_Single(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	server := RunNatsServer(t)
	defer server.Shutdown()

	url := "127.0.0.1"
	testSubject := "test-single"
	testQueue := "test-queue-single"

	config := &config.Config{
		URL:     url,
		Subject: testSubject,
		Queue:   testQueue,
	}

	ns, err := New(config)
	defer ns.Close()
	assert.NoError(t, err)
	assert.NotNil(t, ns)

	nc, err := natslib.Connect(url)
	assert.NoError(t, err)
	defer nc.Close()

	for i := 0; i < 3; i++ {
		err = nc.Publish(testSubject, []byte(fmt.Sprintf("%d", i)))
		assert.NoError(t, err)
	}

	// Prepare a channel to receive messages
	messageCh := make(chan sourcesdk.Message, 10)
	ns.Read(ctx, TestReadRequest{count: 3, timeout: time.Second}, messageCh)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(messageCh))
}

// Test_Multiple tests multiple sources reading from a single nats subject
func Test_Multiple(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	server := RunNatsServer(t)
	defer server.Shutdown()

	url := "127.0.0.1"
	testSubject := "test-single"
	testQueue := "test-queue-single"

	config := &config.Config{
		URL:     url,
		Subject: testSubject,
		Queue:   testQueue,
	}

	ns1, err := New(config)
	defer ns1.Close()
	assert.NoError(t, err)
	assert.NotNil(t, ns1)

	ns2, err := New(config)
	defer ns2.Close()
	assert.NoError(t, err)
	assert.NotNil(t, ns2)

	nc, err := natslib.Connect(url)
	defer nc.Close()
	assert.NoError(t, err)

	for i := 0; i < 5; i++ {
		err = nc.Publish(testSubject, []byte(fmt.Sprintf("%d", i)))
		assert.NoError(t, err)
	}
	messageCh1 := make(chan sourcesdk.Message, 10)
	messageCh2 := make(chan sourcesdk.Message, 10)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			ns1.Read(ctx, TestReadRequest{count: 1, timeout: time.Second}, messageCh1)
			if len(messageCh1)+len(messageCh2) == 5 {
				break
			}
		}
		close(messageCh1)
	}()
	go func() {
		defer wg.Done()
		for {
			ns2.Read(ctx, TestReadRequest{count: 1, timeout: time.Second}, messageCh2)
			if len(messageCh1)+len(messageCh2) == 5 {
				break
			}
		}
		close(messageCh2)
	}()
	wg.Wait()

	var sum int
	for m := range messageCh1 {
		byteToInt, err := strconv.Atoi(string(m.Value()))
		assert.NoError(t, err)
		sum += byteToInt
	}
	for m := range messageCh2 {
		byteToInt, err := strconv.Atoi(string(m.Value()))
		assert.NoError(t, err)
		sum += byteToInt
	}
	assert.Equal(t, 10, sum)
}

// RunNatsServer starts a nats server
func RunNatsServer(t *testing.T) *server.Server {
	t.Helper()
	opts := natstestserver.DefaultTestOptions
	return natstestserver.RunServer(&opts)
}
