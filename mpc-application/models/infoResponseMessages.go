package models

import (
	"encoding/json"

	"mpc_poc/messaging"

	"github.com/koteld/multi-party-sig/pkg/party"
)

type (
	ConfigMessage struct {
		Address   string        `json:"address"`
		IDs       party.IDSlice `json:"participants"`
		SessionID string        `json:"sessionId"`
	}
)

type (
	InfoResponseMessage struct {
		Info    Info            `json:"info"`
		Online  bool            `json:"online"`
		Configs []ConfigMessage `json:"configs"`
	}
)

var infoResponseMessageInputChannels = make(map[party.ID]<-chan *InfoResponseMessage)
var infoResponseMessageOutputChannels = make(map[party.ID]chan<- *InfoResponseMessage)

func GetInfoResponseMessageInputChannel(ID party.ID) <-chan *InfoResponseMessage {
	if infoResponseMessageInputChannels[ID] == nil {
		rawInput := messaging.GetInputChannel(messaging.InfoResponseMessagesChannel + ":" + string(ID))
		res := make(chan *InfoResponseMessage)

		go func() {
			for val := range rawInput {
				bs := &InfoResponseMessage{}
				err := json.Unmarshal(val, bs)
				if err == nil {
					res <- bs
				}
			}
		}()

		infoResponseMessageInputChannels[ID] = res
	}
	return infoResponseMessageInputChannels[ID]
}

func GetInfoResponseMessageOutputChannel(ID party.ID) chan<- *InfoResponseMessage {
	if infoResponseMessageOutputChannels[ID] == nil {
		rawOutput := messaging.GetOutputChannel(messaging.InfoResponseMessagesChannel + ":" + string(ID))
		res := make(chan *InfoResponseMessage)

		go func() {
			for bs := range res {
				val, err := json.Marshal(bs)
				if err == nil {
					rawOutput <- val
				}
			}
		}()

		infoResponseMessageOutputChannels[ID] = res
	}
	return infoResponseMessageOutputChannels[ID]
}
