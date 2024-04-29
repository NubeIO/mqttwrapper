package mqttwrapper

import (
	"encoding/json"
	"fmt"
	"time"
)

type Response struct {
	Body        []byte
	RequestUUID string
	Error       string
}

func (receiver *Response) AsString() string {
	return string(receiver.Body)
}

// RequestResponse sends a request and waits for a response with a timeout
func (m *Mqtt5) RequestResponse(timeoutSeconds int, publishTopic, responseTopic, requestUUID string, body interface{}) *Response {
	respChan := make(chan *Response)
	defer close(respChan)

	// Handle function to receive the response
	handleFunc := func(topic string, payload []byte) {
		resp := &Response{
			Body:        payload,
			RequestUUID: requestUUID,
		}
		respChan <- resp
	}

	// Subscribe to the response topic temporarily
	m.Subscribe(responseTopic, handleFunc)
	defer m.Unsubscribe(responseTopic)

	// Marshal body to JSON for the payload
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return &Response{
			Error:       fmt.Sprintf("Error marshalling body: %v", err),
			RequestUUID: requestUUID,
		}
	}

	// Publish the request
	if err := m.Publish(publishTopic, jsonBody); err != nil {
		return &Response{
			Error:       fmt.Sprintf("Publish failed: %v", err),
			RequestUUID: requestUUID,
		}
	}

	// Wait for the response or timeout
	select {
	case resp := <-respChan:
		return resp
	case <-time.After(time.Duration(timeoutSeconds) * time.Second):
		return &Response{
			Error:       "Response timed out",
			RequestUUID: requestUUID,
		}
	}
}
