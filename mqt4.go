package mqttwrapper

import (
	"crypto/tls"
	"errors"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Mqtt4 struct {
	client mqtt.Client
	opts   *mqtt.ClientOptions
}

func (c *Mqtt4) RequestResponseStream(bufferDuration int, publishTopic, responseTopic, requestUUID string, body interface{}) []*Response {
	//TODO implement me
	panic("implement me")
}

func (c *Mqtt4) RequestResponse(timeoutSeconds int, publishTopic, responseTopic, requestUUID string, body interface{}) *Response {
	//TODO implement me
	panic("implement me")
}

func (c *Mqtt4) Unsubscribe(topics string) error {
	//TODO implement me
	panic("implement me")
}

func (c *Mqtt4) UnsubscribeMany(topics []string) error {
	//TODO implement me
	panic("implement me")
}

func (c *Mqtt4) Publish(topic string, payload interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (c *Mqtt4) StartPublishRateLimiting() {
	//TODO implement me
	panic("implement me")
}

func (c *Mqtt4) StartProcessingMessages() {
	//TODO implement me
	panic("implement me")
}

func (c *Mqtt4) Connect() error {
	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	if !c.client.IsConnectionOpen() {
		return errors.New("not connected")
	}
	return nil
}

func (c *Mqtt4) PublishFull(topic string, payload []byte, qos uint8, retain bool) error {
	token := c.client.Publish(topic, qos, retain, payload)
	return token.Error()
}

func (c *Mqtt4) Subscribe(topic string, fnc SubscribeHandleFunction) error {
	token := c.client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		fnc(msg.Topic(), msg.Payload())
	})
	return token.Error()
}

func connectMQTT4(config Config) (*Mqtt4, error) {
	returnValue := &Mqtt4{}

	returnValue.opts = mqtt.NewClientOptions()
	schema := "tcp"
	if config.Port == 8883 {
		schema = "ssl"
	}
	returnValue.opts.AddBroker(fmt.Sprintf("%s://%s:%d", schema, config.IP, config.Port))
	returnValue.opts.SetClientID(config.Username)
	returnValue.opts.SetUsername(config.Username)
	returnValue.opts.SetPassword(config.Password)
	returnValue.opts.SetProtocolVersion(uint(config.Version))
	returnValue.opts.SetAutoReconnect(true)

	if config.Port == 8883 {
		tlsConfig := tls.Config{}
		tlsConfig.CipherSuites = cipherSuite
		tlsConfig.InsecureSkipVerify = true
		returnValue.opts.SetTLSConfig(&tlsConfig)
	}

	returnValue.opts.AutoReconnect = true

	returnValue.client = mqtt.NewClient(returnValue.opts)

	return returnValue, nil
}
