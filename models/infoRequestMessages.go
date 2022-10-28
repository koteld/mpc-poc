package models

import (
	"encoding/json"

	"mpc_poc/messaging"

	"github.com/koteld/multi-party-sig/pkg/party"
)

type Info string

const (
	Online  Info = "info/online"
	Configs Info = "info/configs"
)

type (
	InfoRequestMessage struct {
		Info Info `json:"info"`
	}
)

var infoRequestMessageInputChannels = make(map[party.ID]<-chan *InfoRequestMessage)
var infoRequestMessageOutputChannels = make(map[party.ID]chan<- *InfoRequestMessage)

func GetInfoRequestMessageInputChannel(ID party.ID) <-chan *InfoRequestMessage {
	if infoRequestMessageInputChannels[ID] == nil {
		rawInput := messaging.GetInputChannel(messaging.InfoRequestMessagesChannel + ":" + string(ID))
		res := make(chan *InfoRequestMessage)

		go func() {
			for val := range rawInput {
				bs := &InfoRequestMessage{}
				err := json.Unmarshal(val, bs)
				if err == nil {
					res <- bs
				}
			}
		}()

		infoRequestMessageInputChannels[ID] = res
	}
	return infoRequestMessageInputChannels[ID]
}

func GetInfoRequestMessageOutputChannel(ID party.ID) chan<- *InfoRequestMessage {
	if infoRequestMessageOutputChannels[ID] == nil {
		rawOutput := messaging.GetOutputChannel(messaging.InfoRequestMessagesChannel + ":" + string(ID))
		res := make(chan *InfoRequestMessage)

		go func() {
			for bs := range res {
				val, err := json.Marshal(bs)
				if err == nil {
					rawOutput <- val
				}
			}
		}()

		infoRequestMessageOutputChannels[ID] = res
	}
	return infoRequestMessageOutputChannels[ID]
}
