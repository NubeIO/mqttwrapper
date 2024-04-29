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
	Timestamp   string
}

func (receiver *Response) GetBody() []byte {
	return receiver.Body
}

func (receiver *Response) AsString() string {
	if receiver.Error != "" {
		return fmt.Sprintf("err: %s uuid: %s", receiver.Error, receiver.RequestUUID)
	}
	return string(receiver.Body)
}

func (receiver *Response) GetUUID() string {
	return receiver.RequestUUID
}

// RequestResponseStream sends a request and buffers all responses for the specified duration in seconds
func (m *Mqtt5) RequestResponseStream(bufferDuration int, publishTopic, responseTopic, requestUUID string, body interface{}) []*Response {
	var responses []*Response
	respChan := make(chan *Response, 100) // Buffered channel to avoid blocking senders

	// handle incoming messages
	handleFunc := func(topic string, payload []byte) {
		resp := &Response{
			Body:        payload,
			RequestUUID: requestUUID,
			Timestamp:   time.Now().Format(time.RFC3339),
		}
		respChan <- resp
	}

	if err := m.Subscribe(responseTopic, handleFunc); err != nil {
		responses = append(responses, &Response{
			Error:       fmt.Sprintf("Subscribe failed: %v", err),
			RequestUUID: requestUUID,
		})
		return responses
	}
	defer m.Unsubscribe(responseTopic)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		responses = append(responses, &Response{
			Error:       fmt.Sprintf("Error marshalling body: %v", err),
			RequestUUID: requestUUID,
		})
		return responses
	}

	// Publish the request
	if err := m.Publish(publishTopic, jsonBody); err != nil {
		responses = append(responses, &Response{
			Error:       fmt.Sprintf("Publish failed: %v", err),
			RequestUUID: requestUUID,
		})
		return responses
	}

	// Collect responses until the buffer duration expires
	bufferTimer := time.NewTimer(time.Duration(bufferDuration) * time.Second)
	defer bufferTimer.Stop()

	collecting := true
	go func() {
		<-bufferTimer.C
		close(respChan) // Close channel to signal no more messages will be sent
		collecting = false
	}()

	for collecting {
		select {
		case resp, ok := <-respChan:
			if ok {
				responses = append(responses, resp)
			}
		default:
			time.Sleep(50 * time.Millisecond) // Reduce CPU usage
		}
	}

	return responses
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
			Timestamp:   time.Now().Format(time.RFC3339),
		}
		respChan <- resp
	}

	// Subscribe to the response topic temporarily
	err := m.Subscribe(responseTopic, handleFunc)
	if err != nil {
		return &Response{
			Error:       fmt.Sprintf("Subscribe failed: %v", err),
			RequestUUID: requestUUID,
		}
	}
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
