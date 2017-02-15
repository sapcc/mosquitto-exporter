package main

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	client    MQTT.Client
	Endpoints []string
}

func mqttClient(c *cli.Context) (*MQTTClient, error) {
	// check options
	endpoints := []string{}
	if len(c.StringSlice("endpoint")) > 0 {
		endpoints = c.StringSlice("endpoint")
	} else {
		return nil, fmt.Errorf("No transport endpoints given")
	}

	// create options
	opts := MQTT.NewClientOptions()
	for _, endpoint := range c.StringSlice("endpoint") {
		logrus.Info("Using MQTT broker ", endpoint)
		opts.AddBroker(endpoint)
	}
	opts.SetCleanSession(true)
	return &MQTTClient{
		Endpoints: endpoints,
		client:    MQTT.NewClient(opts),
	}, nil
}

func (c *MQTTClient) Subscribe(topic string) (<-chan MQTT.Message, func()) {
	msgChan := make(chan MQTT.Message)
	canceled := make(chan struct{})
	var mutex sync.Mutex

	messageCallback := func(mClient MQTT.Client, mMessage MQTT.Message) {
		mutex.Lock()
		select {
		case <-canceled:
		case msgChan <- mMessage:
		}
		mutex.Unlock()
	}

	cancel := func() {
		c.unsubscribe(topic)
		close(canceled)
		mutex.Lock()
		close(msgChan)
		mutex.Unlock()
	}

	c.subscribe(topic, 0, messageCallback)
	return msgChan, cancel
}

func (c *MQTTClient) Connect() error {
	logrus.Info("Connecting to MQTT broker")
	token := c.client.Connect()
	if !token.WaitTimeout(10 * time.Second) {
		return errors.New("Timeout connecting to broker")
	}
	return token.Error()
}

func (c *MQTTClient) Disconnect() {
	c.client.Disconnect(1000)
}

func (c *MQTTClient) subscribe(topic string, qos byte, cb MQTT.MessageHandler) {
	c.client.Subscribe(topic, 0, cb).Wait()
}

func (c *MQTTClient) unsubscribe(topic string) bool {
	return c.client.Unsubscribe(topic).WaitTimeout(500 * time.Millisecond)
}
