package mqtt

import (
	"log"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttClient interface {
	Publish(string, string)
	RegisterTopic(string, mqtt.MessageHandler)
	DeregisterTopic(string)
}

type Topic interface{}

type topic struct {
	handler mqtt.MessageHandler
	path    string
}

type Message struct {
	topic   Topic
	message string
}

type Config struct {
	exporter   mqttClient
	connection Connection
	id         string
}

type Connection interface {
	Connect(string, string, int)
}

type connection struct {
	url  string
	port int
}

type mqttClient struct {
	client      mqtt.Client
	topics      map[string]mqtt.MessageHandler
	writeMutex  *sync.Mutex
	topicsMutex *sync.RWMutex
}

func (c *mqttClient) Publish(topic string, message string) error {
	c.writeMutex.Lock()
	token := c.client.Publish(topic, 1, false, message)
	c.writeMutex.Unlock()
	if token.WaitTimeout(time.Second*5) && token.Error() != nil {
		log.Printf("error in publishing a message: %s", token.Error())
	}
	return nil
}

func (c *mqttClient) RegisterTopic(topic string, handler mqtt.MessageHandler) {
	c.topicsMutex.Lock()
	defer c.topicsMutex.Unlock()
	c.topics[topic] = handler
	token := c.client.Subscribe(topic, 1, handler)
	if token.WaitTimeout(time.Second*5) && token.Error() != nil {
		log.Printf("error in register topic: %s", token.Error())
	}
}

func (c *mqttClient) DeregisterTopic(topic string) {
	c.topicsMutex.Lock()
	defer c.topicsMutex.Unlock()
	token := c.client.Unsubscribe(topic)
	delete(c.topics, topic)
	if token.WaitTimeout(time.Second*5) && token.Error() != nil {
		log.Printf("error in deregister topic: %s", token.Error())
	}
}
