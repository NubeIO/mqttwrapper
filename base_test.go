package mqttwrapper

import (
	"fmt"
	"testing"
	"time"
)

func TestNewMqttClient(t *testing.T) {
	cli, err := NewMqttClient(Config{
		IP:                     "localhost",
		Port:                   1883,
		Version:                V5,
		Username:               "",
		Password:               "",
		SubscribeRateMax:       100,
		SubscribeRatePeriodSec: 100000,
		PublishRateMax:         100,
		PublishRatePeriodSec:   100000,
		QoS:                    0,
		Retain:                 false,
	})
	fmt.Println(err)

	err = cli.Connect()
	fmt.Println(err)

	if err != nil {
		return
	}
	cli.StartProcessingMessages()

	cli.Subscribe("ros/global/+", handleFunction)

	err = cli.Publish("ros/global/1234", "1234ssss")
	err = cli.Publish("ros/hey", "1234")
	err = cli.Publish("ros/global", "1234")

	fmt.Println(err)
	if err != nil {
		return
	}
	time.Sleep(2 * time.Second)

}

func handleFunction(topic string, body []byte) {

}

func TestRequestResponse(t *testing.T) {
	cli, err := NewMqttClient(Config{
		IP:                     "localhost",
		Port:                   1883,
		Version:                V5,
		Username:               "",
		Password:               "",
		SubscribeRateMax:       100,
		SubscribeRatePeriodSec: 100000,
		PublishRateMax:         100,
		PublishRatePeriodSec:   100000,
		QoS:                    0,
		Retain:                 false,
	})
	fmt.Println(err)

	err = cli.Connect()
	fmt.Println(err)

	if err != nil {
		return
	}
	cli.StartProcessingMessages()
	resp := cli.RequestResponse(10, "test", "test/resp", "abc", "hello")
	fmt.Println(string(resp.Body))
}
