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
	cli.StartPublishRateLimiting()
	cli.StartProcessingMessages()

	cli.Subscribe("ros/global/+", handleFunction)

	err = cli.Publish("ros/global/1234", "1234")
	err = cli.Publish("ros/hey", "1234")
	err = cli.Publish("ros/global", "1234")

	fmt.Println(err)
	if err != nil {
		return
	}
	time.Sleep(2 * time.Second)

}

func handleFunction(uuid string, payload []byte) {
	fmt.Println(11111)
	fmt.Printf("Received message on %s: %s\n", uuid, string(payload))
	fmt.Printf("Received on [%s]:", uuid)

	//mq := mqparse.NewMQTTParse(topic, payload)

	//err := mq.ParsePayload()
	//fmt.Println(err, len(mq.GetAllFuncs()))
	//pprint.PrintJSON(mq.GetAllFuncs())
	//
	//if mq.TopicPartExists("whois") {
	//	//resp, err := inst.whoIs()
	//	//if err != nil {
	//	//	return
	//	//}
	//	//marshal, err := json.Marshal(resp)
	//	//if err != nil {
	//	//	return
	//	//}
	//	//inst.mqttPublish("ros/global/iam", marshal)
	//}

}
