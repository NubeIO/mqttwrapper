package mqttwrapper

import (
	"fmt"
	"testing"
)

func TestMqtt5_RequestResponseStream(t *testing.T) {
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
	resp := cli.RequestResponseStream(10, "test", "r/res/v1/cloud/RX-2/plain/command/R-1/req-uuid", "abc", "hello")
	for i, response := range resp {
		fmt.Println(i, response.AsString())
	}
}
