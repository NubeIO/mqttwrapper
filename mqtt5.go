package mqttwrapper

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Mqtt5 struct {
	cm                     *autopaho.ConnectionManager
	cfg                    autopaho.ClientConfig
	fncMap                 map[string]SubscribeHandleFunction
	config                 Config
	msgChan                chan *paho.Publish
	publishMsgChan         chan *publishMessage
	wg                     sync.WaitGroup
	SubscribeRateMax       int
	SubscribeRatePeriodSec time.Duration // Time period for rate limiting
	PublishRateMax         int           // Max number of messages per time period for publishing
	PublishRatePeriodSec   time.Duration
	qos                    uint8
	retain                 bool
	connectionUp           bool
}

// TopicMatch checks if the incoming topic matches the subscription pattern
// considering MQTT wildcards + and #.
func TopicMatch(subPattern, topic string) bool {
	// Split patterns and topics
	patternParts := strings.Split(subPattern, "/")
	topicParts := strings.Split(topic, "/")

	for i, part := range patternParts {
		if part == "#" {
			return true // # wildcard matches all remaining topics
		}
		if i >= len(topicParts) {
			return false // Topic is shorter than the pattern
		}
		if part != "+" && part != topicParts[i] {
			return false // Exact match required and not met
		}
	}

	// If we're here, all parts matched. Also check lengths to be sure.
	return len(patternParts) == len(topicParts)
}

func (m *Mqtt5) messageHandler(msg *paho.Publish) {
	// Modified to use TopicMatch for wildcard support
	for pattern, _ := range m.fncMap {
		if TopicMatch(pattern, msg.Topic) {
			select {
			case m.msgChan <- msg:
				// Call the handler in your processing loop, not here, to avoid blocking
			default:
				log.Print("DUMP MESSAGES due to full channel")
			}
			// Break if you assume one handler per message or continue if multiple handlers are allowed
			break
		}
	}
}
func (m *Mqtt5) Close() {
	close(m.msgChan)
	m.wg.Wait() // Wait for all message processing to finish
}

func (m *Mqtt5) StartProcessingMessages() {
	m.wg.Add(1)
	m.msgChan = make(chan *paho.Publish, m.SubscribeRateMax)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(m.SubscribeRatePeriodSec * time.Second)
		defer ticker.Stop()
		msgCount := 0
		for {
			select {
			case msg := <-m.msgChan:
				log.Printf("Received message on channel for topic %s", msg.Topic)
				// Reset msgCount based on your rate limiting logic
				//if msgCount < m.SubscribeRateMax {
				// Check for a matching handler using the TopicMatch function
				for pattern, handler := range m.fncMap {
					if TopicMatch(pattern, msg.Topic) {
						handler(msg.Topic, msg.Payload)
						msgCount++
						break // Assume one handler per message; remove if multiple handlers are allowed
					}
				}
				//} else {
				//	log.Println("Dropping message due to rate limit")
				//}
			case <-ticker.C:
				msgCount = 0
			}
		}
	}()
}

func (m *Mqtt5) Connect() error {
	var err error
	m.cm, err = autopaho.NewConnection(context.Background(), m.cfg)
	if err != nil {
		return err
	}

	err = m.cm.AwaitConnection(context.Background())
	if err != nil {
		return err
	}

	return nil
}

type publishMessage struct {
	topic   string
	payload []byte
	qos     uint8
	retain  bool
}

func (m *Mqtt5) startPublishRateLimiting() {
	m.publishMsgChan = make(chan *publishMessage, m.PublishRateMax)
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		ticker := time.NewTicker(m.PublishRatePeriodSec * time.Second)
		defer ticker.Stop()

		msgCount := 0

		for {
			select {
			case msg := <-m.publishMsgChan:
				//if msgCount < m.PublishRateMax {
				// Publish message
				_, err := m.cm.Publish(context.Background(), &paho.Publish{
					Topic:   msg.topic,
					Payload: msg.payload,
					QoS:     msg.qos,
					Retain:  msg.retain,
				})
				if err != nil {
					log.Printf("Error publishing message: %v", err)
				}
				msgCount++
				//} else {
				//	log.Println("Dropping publish message due to rate limit")
				//}
			case <-ticker.C:
				msgCount = 0
			}
		}
	}()
}

func (m *Mqtt5) PublishFull(topic string, payload []byte, qos uint8, retain bool) error {
	select {
	case m.publishMsgChan <- &publishMessage{topic, payload, qos, retain}:
		return nil
	default:
		return errors.New("publish message dropped due to rate limit")
	}
}

func (m *Mqtt5) Publish(topic string, payload interface{}) error {
	if m.cm == nil {
		return errors.New("cant not add subscription as there is no connection to the broker")
	}
	if m.isConnectDown() {
		return errors.New("no current connection to the broker")
	}
	p, err := m.checkPayload(payload)
	if err != nil {
		return err
	}
	_, err = m.cm.Publish(context.Background(), &paho.Publish{
		Topic:   topic,
		Payload: p,
		QoS:     m.qos,
		Retain:  m.retain,
	})
	return err

}

func (m *Mqtt5) checkPayload(payload interface{}) ([]byte, error) {
	switch p := payload.(type) {
	case string:
		return []byte(p), nil
	case []byte:
		return p, nil
	case bytes.Buffer:
		return p.Bytes(), nil
	default:
		return nil, errors.New("unknown payload type")
	}
}

func (m *Mqtt5) UnsubscribeMany(topics []string) error {
	if m.cm == nil {
		return errors.New("cant not add subscription as there is no connection to the broker")
	}
	if m.isConnectDown() {
		return errors.New("no current connection to the broker")
	}
	u := &paho.Unsubscribe{
		Topics:     topics,
		Properties: nil,
	}
	_, err := m.cm.Unsubscribe(context.Background(), u)
	return err
}

func (m *Mqtt5) Unsubscribe(topic string) error {
	if m.cm == nil {
		return errors.New("cant not add subscription as there is no connection to the broker")
	}
	if m.isConnectDown() {
		return errors.New("no current connection to the broker")
	}
	u := &paho.Unsubscribe{
		Topics:     []string{topic},
		Properties: nil,
	}
	_, err := m.cm.Unsubscribe(context.Background(), u)
	return err
}

func (m *Mqtt5) Subscribe(topic string, fnc SubscribeHandleFunction) error {
	if m.cm == nil {
		return errors.New("cant not add subscription as there is no connection to the broker")
	}
	if m.isConnectDown() {
		return errors.New("no current connection to the broker")
	}
	if m.fncMap == nil {
		m.fncMap = map[string]SubscribeHandleFunction{}
	}

	if _, ok := m.fncMap[topic]; ok {
		return errors.New("topic already subscribed")
	}
	m.fncMap[topic] = fnc

	var topics []paho.SubscribeOptions
	newTopic := paho.SubscribeOptions{
		Topic: topic,
	}
	topics = append(topics, newTopic)
	_, err := m.cm.Subscribe(context.Background(), &paho.Subscribe{
		Properties:    nil,
		Subscriptions: topics,
	})

	if err != nil {

	}

	return nil
}

func (m *Mqtt5) isConnectUp() bool {
	return m.connectionUp
}
func (m *Mqtt5) isConnectDown() bool {
	if m.connectionUp {
		return false
	}
	return true
}

func (m *Mqtt5) onConnectError(err error) {
	m.connectionUp = false
}

func (m *Mqtt5) onConnectUp(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
	m.connectionUp = true
}

func connectMQTT5(config Config) (*Mqtt5, error) {
	inst := &Mqtt5{
		SubscribeRateMax:       config.SubscribeRateMax,
		SubscribeRatePeriodSec: config.SubscribeRatePeriodSec,
		PublishRateMax:         config.PublishRateMax,
		PublishRatePeriodSec:   config.PublishRatePeriodSec,
		qos:                    config.QoS,
	}

	schema := "mqtt"
	if config.Port == 8883 {
		schema = "ssl"
	}
	inst.cfg = autopaho.ClientConfig{
		BrokerUrls: []*url.URL{
			{
				Scheme: schema,
				Host:   fmt.Sprintf("%s:%d", config.IP, config.Port),
			},
		},
		TlsCfg: &tls.Config{
			CipherSuites:       cipherSuite,
			InsecureSkipVerify: true,
		},
		OnConnectionUp: inst.onConnectUp,
		OnConnectError: inst.onConnectError,
		ClientConfig: paho.ClientConfig{
			ClientID: config.Username,
			Router:   paho.NewSingleHandlerRouter(inst.messageHandler),
		},
	}

	return inst, nil
}
