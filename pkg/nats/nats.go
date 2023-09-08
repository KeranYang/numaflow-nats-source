package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	natslib "github.com/nats-io/nats.go"
	sourcesdk "github.com/numaproj/numaflow-go/pkg/sourcer"
	"go.uber.org/zap"

	"numaflow-nats-source/pkg/config"
	"numaflow-nats-source/pkg/utils"
)

const (
	defaultBufferSize = 1000
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

	volumeReader utils.VolumeReader

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

func New(c *config.Config, opts ...Option) (*natsSource, error) {
	n := &natsSource{
		bufferSize: defaultBufferSize,
	}
	for _, o := range opts {
		if err := o(n); err != nil {
			return nil, err
		}
	}
	if n.logger == nil {
		var err error
		n.logger, err = zap.NewDevelopment()
		if err != nil {
			return nil, fmt.Errorf("failed to create logger, %w", err)
		}
	}

	n.messages = make(chan *Message, n.bufferSize)
	n.volumeReader = utils.NewNatsVolumeReader(utils.SecretVolumePath)

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

	if c.TLS != nil {
		if c, err := utils.GetTLSConfig(c.TLS, n.volumeReader); err != nil {
			return nil, err
		} else {
			opt = append(opt, natslib.Secure(c))
		}
	}

	if c.Auth != nil {
		switch {
		case c.Auth.Basic != nil && c.Auth.Basic.User != nil && c.Auth.Basic.Password != nil:
			username, err := n.volumeReader.GetSecretFromVolume(c.Auth.Basic.User)
			if err != nil {
				return nil, fmt.Errorf("failed to get basic auth user, %w", err)
			}
			password, err := n.volumeReader.GetSecretFromVolume(c.Auth.Basic.Password)
			if err != nil {
				return nil, fmt.Errorf("failed to get basic auth password, %w", err)
			}
			opt = append(opt, natslib.UserInfo(username, password))
		case c.Auth.Token != nil:
			token, err := n.volumeReader.GetSecretFromVolume(c.Auth.Token)
			if err != nil {
				return nil, fmt.Errorf("failed to get auth token, %w", err)
			}
			opt = append(opt, natslib.Token(token))
		case c.Auth.NKey != nil:
			nKeyFile, err := n.volumeReader.GetSecretVolumePath(c.Auth.NKey)
			if err != nil {
				return nil, fmt.Errorf("failed to get configured nkey file, %w", err)
			}
			o, err := natslib.NkeyOptionFromSeed(nKeyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to get NKey, %w", err)
			}
			opt = append(opt, o)
		}
	}

	n.logger.Info("Connecting to nats service...")
	if conn, err := natslib.Connect(c.URL, opt...); err != nil {
		n.logger.Error("Failed to connect to nats server", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to nats server, %w", err)
	} else {
		n.natsConn = conn
	}

	n.logger.Info(fmt.Sprintf("Subscribing to subject %s with queue %s", c.Subject, c.Queue))
	if sub, err := n.natsConn.QueueSubscribe(c.Subject, c.Queue, func(msg *natslib.Msg) {
		readOffset := uuid.New().String()
		m := &Message{
			payload:    string(msg.Data),
			readOffset: readOffset,
			id:         readOffset,
		}
		n.messages <- m
	}); err != nil {
		n.logger.Error("Failed to QueueSubscribe nats messages", zap.Error(err))
		n.natsConn.Close()
		return nil, fmt.Errorf("failed to QueueSubscribe nats messages, %w", err)
	} else {
		n.sub = sub
	}
	n.logger.Info("Nats source server started")
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
	// Ack is a no-op for the Nats source.
}

func (n *natsSource) Close() error {
	n.logger.Info("Shutting down nats source server...")
	if err := n.sub.Unsubscribe(); err != nil {
		n.logger.Error("Failed to unsubscribe nats subscription", zap.Error(err))
	}
	n.natsConn.Close()
	n.logger.Info("Nats source server shutdown")
	return nil
}
