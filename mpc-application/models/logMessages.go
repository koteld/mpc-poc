package models

import (
	"encoding/json"

	"mpc_poc/messaging"
)

type (
	LogMessage struct {
		SessionID   string   `json:"sessionID"`
		Participant string   `json:"participant"`
		Protocol    Protocol `json:"protocol"`
		Round       uint16   `json:"round"`
		Message     string   `json:"message"`
		Timestamp   string   `json:"timestamp"`
		IP          string   `json:"ip"`
	}
)

func GetLogMessageInputChannel() <-chan *LogMessage {
	rawInput := messaging.GetInputChannel(messaging.LogMessagesChannel)
	res := make(chan *LogMessage)

	go func() {
		for val := range rawInput {
			bs := &LogMessage{}
			err := json.Unmarshal(val, bs)
			if err == nil {
				res <- bs
			}
		}
	}()

	return res
}

func GetLogMessageOutputChannel() chan<- *LogMessage {
	rawOutput := messaging.GetOutputChannel(messaging.LogMessagesChannel)
	res := make(chan *LogMessage)

	go func() {
		for bs := range res {
			val, err := json.Marshal(bs)
			if err == nil {
				rawOutput <- val
			}
		}
	}()

	return res
}
