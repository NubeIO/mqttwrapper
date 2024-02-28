package mqttwrapper

import (
	"crypto/tls"
	"errors"
	"time"
)

var cipherSuite []uint16 = []uint16{
	tls.TLS_AES_128_GCM_SHA256,
	tls.TLS_AES_256_GCM_SHA384,
	tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_RSA_WITH_RC4_128_SHA,
}

type Config struct {
	IP                     string
	Port                   uint
	Version                Version
	Username               string
	Password               string
	SubscribeRateMax       int
	SubscribeRatePeriodSec time.Duration
	PublishRateMax         int // Max number of messages per time period for publishing
	PublishRatePeriodSec   time.Duration
	QoS                    uint8
	Retain                 bool
}

type SubscribeHandleFunction func(topic string, payload []byte)

type MQTT interface {
	Connect() error
	Publish(topic string, payload interface{}) error
	PublishFull(topic string, payload []byte, qos uint8, retain bool) error
	Subscribe(topic string, fnc SubscribeHandleFunction) error
	Unsubscribe(topics string) error
	UnsubscribeMany(topics []string) error
	StartProcessingMessages()
	StartPublishRateLimiting()
}

// Version of the cli
type Version int

const (
	// V3 is MQTT Version 3
	V3 Version = iota
	// V5 is MQTT Version 5
	V5
)

// ConnectionState of the Client
type ConnectionState bool

const (
	// Disconnected : no connection to broker
	Disconnected ConnectionState = false
	// Connected : connection established to broker
	Connected = true
)

func NewMqttClient(config Config) (MQTT, error) {

	if config.IP == "" {
		config.IP = "localhost"
	}
	if config.Port == 0 {
		config.Port = 1883
	}
	if config.Version != V5 {
		config.Version = V5
	}
	if config.SubscribeRateMax == 0 {
		config.SubscribeRateMax = 100
	}

	if config.SubscribeRatePeriodSec == 0 {
		config.SubscribeRatePeriodSec = 30
	}

	if config.PublishRateMax == 0 {
		config.PublishRateMax = 100
	}
	if config.PublishRatePeriodSec == 0 {
		config.PublishRatePeriodSec = 30
	}

	if config.QoS > 2 || config.QoS < 0 {
		return nil, errors.New("value of qos must be 0, 1, 2")
	}

	if config.Version >= V5 {
		return connectMQTT5(config)
	} else {
		return connectMQTT4(config)
	}
}
