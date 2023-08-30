package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	natslib "github.com/nats-io/nats.go"
	sourcesdk "github.com/numaproj/numaflow-go/pkg/sourcer"
	"go.uber.org/zap"

	"numaflow-nats-source/pkg/configuration"
	"numaflow-nats-source/pkg/utils"
)

type Message struct {
	payload    string
	readOffset string
	id         string
}

type natsSource struct {
	natsConn *natslib.Conn
	sub      *natslib.Subscription

	bufferSize int
	messages   chan *Message

	logger *zap.Logger
}

type Option func(*natsSource) error

// WithLogger is used to return logger information
func WithLogger(l *zap.Logger) Option {
	return func(o *natsSource) error {
		o.logger = l
		return nil
	}
}

func New(c *configuration.Config, opts ...Option) (*natsSource, error) {
	n := &natsSource{
		bufferSize: 1000, // default size
	}
	for _, o := range opts {
		if err := o(n); err != nil {
			return nil, err
		}
	}
	if n.logger == nil {
		n.logger, _ = zap.NewDevelopment()
	}

	n.messages = make(chan *Message, n.bufferSize)

	opt := []natslib.Option{
		natslib.MaxReconnects(-1),
		natslib.ReconnectWait(3 * time.Second),
		natslib.DisconnectHandler(func(c *natslib.Conn) {
			n.logger.Info("Nats disconnected")
		}),
		natslib.ReconnectHandler(func(c *natslib.Conn) {
			n.logger.Info("Nats reconnected")
		}),
	}

	if c.Auth != nil && c.Auth.Token != nil {
		token, err := utils.GetSecretFromVolume(c.Auth.Token)
		if err != nil {
			return nil, fmt.Errorf("failed to get auth token, %w", err)
		}
		opt = append(opt, natslib.Token(token))
	}

	n.logger.Info("Connecting to nats service...")
	if conn, err := natslib.Connect(c.URL, opt...); err != nil {
		return nil, fmt.Errorf("failed to connect to nats server, %w", err)
	} else {
		n.natsConn = conn
	}

	if sub, err := n.natsConn.QueueSubscribe(c.Subject, c.Queue, func(msg *natslib.Msg) {
		readOffset := uuid.New().String()
		m := &Message{
			payload:    string(msg.Data),
			readOffset: readOffset,
			id:         readOffset,
		}
		n.messages <- m
	}); err != nil {
		n.natsConn.Close()
		return nil, fmt.Errorf("failed to QueueSubscribe nats messages, %w", err)
	} else {
		n.sub = sub
	}
	return n, nil
}

// Pending returns the number of pending records.
func (n *natsSource) Pending(_ context.Context) uint64 {
	// The nats source always returns zero to indicate no pending records.
	return 0
}

func (n *natsSource) Read(_ context.Context, readRequest sourcesdk.ReadRequest, messageCh chan<- sourcesdk.Message) {
	// Handle the timeout specification in the read request.
	ctx, cancel := context.WithTimeout(context.Background(), readRequest.TimeOut())
	defer cancel()

	// Read the data from the source and send the data to the message channel.
	for i := 0; uint64(i) < readRequest.Count(); i++ {
		select {
		case <-ctx.Done():
			// If the context is done, the read request is timed out.
			return
		case m := <-n.messages:
			// Otherwise, we read the data from the source and send the data to the message channel.
			messageCh <- sourcesdk.NewMessage(
				[]byte(m.payload),
				sourcesdk.NewOffset([]byte(m.readOffset), "0"),
				time.Now())
		}
	}
}

// Ack acknowledges the data from the source.
func (n *natsSource) Ack(_ context.Context, request sourcesdk.AckRequest) {
	for _, offset := range request.Offsets() {
		n.logger.Info(fmt.Sprintf("Acking offset %s", string(offset.Value())))
	}
}
