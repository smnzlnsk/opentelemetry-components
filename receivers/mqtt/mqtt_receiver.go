package mqttreceiver // import github.com/smnzlnsk/opentelemetry-components/receivers/mqtt

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

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

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	receiver := &mqttReceiver{
		config:      cfg,
		client:      client,
		logger:      logger,
		consumer:    consumer,
		topics:      make(map[string]mqtt.MessageHandler),
		writeMutex:  &sync.Mutex{},
		topicsMutex: &sync.RWMutex{},
	}

	receiver.RegisterTopic(receiver.config.Topic, receiver.handleMetrics)
	return receiver, nil
}

func (mr *mqttReceiver) Start(ctx context.Context, host component.Host) error {
	ctx = context.Background()
	ctx, mr.cancel = context.WithCancel(ctx)
	marshaler, err := newMarshaler(mr.config.Encoding)
	if err != nil {
		return err
	}
	mr.marshaler = marshaler
	mr.host = host

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (mr *mqttReceiver) Shutdown(ctx context.Context) error {
	if mr.cancel != nil {
		mr.cancel()
	}
	mr.client.Disconnect(250)
	return nil
}

func (mr *mqttReceiver) ConsumeMetrics(ctx context.Context, metrics pmetric.Metrics) error {
	if mr.consumer == nil {
		mr.logger.Error("no next consumer available, dropping metric data")
		return nil
	}

	err := mr.consumer.ConsumeMetrics(ctx, metrics)
	if err != nil {
		mr.logger.Error("failed to forward metric data", zap.Error(err))
		return err
	}
	mr.logger.Debug("successfully consumed metric data")
	return nil
}

func (mr *mqttReceiver) handleMetrics(c mqtt.Client, m mqtt.Message) {
	data, err := mr.marshaler.metricsUnmarshaler.UnmarshalMetrics(m.Payload())
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
		log.Printf("error in register topic: %s", token.Error())
	}
}

func (mr *mqttReceiver) DeregisterTopic(topic string) {
	mr.topicsMutex.Lock()
	defer mr.topicsMutex.Unlock()
	token := mr.client.Unsubscribe(topic)
	delete(mr.topics, topic)
	if token.WaitTimeout(time.Second*5) && token.Error() != nil {
		log.Printf("error in deregister topic: %s", token.Error())
	}
}
