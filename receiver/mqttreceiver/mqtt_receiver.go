package mqttreceiver // import github.com/smnzlnsk/opentelemetry-components/receiver/mqttreceiver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/collector/receiver"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

var _ receiver.Metrics = (*mqttReceiver)(nil)

type mqttReceiver struct {
	config      *Config
	logger      *zap.Logger
	consumer    consumer.Metrics
	marshaler   *marshaler
	client      mqtt.Client
	host        component.Host
	cancel      context.CancelFunc
	topics      map[string]mqtt.MessageHandler
	writeMutex  *sync.Mutex
	topicsMutex *sync.RWMutex
}

func newMQTTReceiver(cfg *Config, logger *zap.Logger, consumer consumer.Metrics) (*mqttReceiver, error) {
	uri := fmt.Sprintf("%s:%d", cfg.Broker.Host, cfg.Broker.Port)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(uri)
	opts.SetClientID(cfg.ClientID)

	// Set connection timeout
	opts.SetConnectTimeout(30 * time.Second)

	// Auto reconnect settings
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(5 * time.Minute)
	opts.SetKeepAlive(30 * time.Second)

	// Set clean session to false for persistent session
	opts.SetCleanSession(false)

	// Set handlers for connection events
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		logger.Warn("MQTT connection lost", zap.Error(err))
	})

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		logger.Info("MQTT connection established")
	})

	opts.SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
		logger.Info("MQTT attempting to reconnect")
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	r := &mqttReceiver{
		config:      cfg,
		client:      client,
		logger:      logger,
		consumer:    consumer,
		topics:      make(map[string]mqtt.MessageHandler),
		writeMutex:  &sync.Mutex{},
		topicsMutex: &sync.RWMutex{},
	}

	return r, nil
}

func (mr *mqttReceiver) Start(ctx context.Context, host component.Host) error {
	_, mr.cancel = context.WithCancel(ctx)
	marshaler, err := newMarshaler(mr.config.Encoding)
	if err != nil {
		return err
	}
	mr.marshaler = marshaler
	mr.host = host

	mr.RegisterTopic(mr.config.Topic, mr.handleMetrics)

	go func() {
		<-ctx.Done()
		mr.Shutdown(ctx)
	}()

	return nil
}

func (mr *mqttReceiver) Shutdown(ctx context.Context) error {
	if mr.cancel != nil {
		mr.cancel()
	}

	// Unsubscribe from all topics before disconnecting
	mr.topicsMutex.RLock()
	for topic := range mr.topics {
		token := mr.client.Unsubscribe(topic)
		token.WaitTimeout(2 * time.Second)
	}
	mr.topicsMutex.RUnlock()

	// Disconnect with a reasonable timeout
	if mr.client.IsConnected() {
		mr.client.Disconnect(1000) // 1 second timeout for disconnect
	}

	return nil
}

func (mr *mqttReceiver) ConsumeMetrics(ctx context.Context, metrics pmetric.Metrics) error {
	if mr.consumer == nil {
		return fmt.Errorf("no consumer available to receive metrics")
	}
	return mr.consumer.ConsumeMetrics(ctx, metrics)
}

func (mr *mqttReceiver) handleMetrics(c mqtt.Client, m mqtt.Message) {
	data, err := mr.marshaler.metricsUnmarshaler.UnmarshalMetrics(m.Payload())
	mr.logger.Debug("received metric data")
	if err != nil {
		mr.logger.Error("could not unmarshal message")
		return
	}
	mr.logger.Debug("successfully unmarshaled message")
	err = mr.ConsumeMetrics(context.Background(), data)
	if err != nil {
		mr.logger.Error("failed to consume metrics", zap.Error(err))
	}
}

func (mr *mqttReceiver) RegisterTopic(topic string, handler mqtt.MessageHandler) {
	mr.topicsMutex.Lock()
	defer mr.topicsMutex.Unlock()
	mr.topics[topic] = handler
	token := mr.client.Subscribe(topic, 1, handler)
	if token.WaitTimeout(time.Second*5) && token.Error() != nil {
		mr.logger.Error("error in register topic: %s", zap.Error(token.Error()))
	}
}

func (mr *mqttReceiver) DeregisterTopic(topic string) {
	mr.topicsMutex.Lock()
	defer mr.topicsMutex.Unlock()
	token := mr.client.Unsubscribe(topic)
	delete(mr.topics, topic)
	if token.WaitTimeout(time.Second*5) && token.Error() != nil {
		mr.logger.Error("error in deregister topic: %s", zap.Error(token.Error()))
	}
}
