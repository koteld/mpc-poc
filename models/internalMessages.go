package models

import (
	"encoding/json"

	"mpc_poc/messaging"

	"github.com/taurusgroup/multi-party-sig/pkg/party"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
)

type (
	InternalMessage struct {
		Message protocol.Message `json:"message"`
	}
)

var internalMessageInputChannels = make(map[party.ID]<-chan *InternalMessage)
var internalMessageOutputChannels = make(map[party.ID]chan<- *InternalMessage)

func GetInternalMessageInputChannel(ID party.ID) <-chan *InternalMessage {
	if internalMessageInputChannels[ID] == nil {
		rawInput := messaging.GetInputChannel(messaging.InternalMessagesChannel + ":" + string(ID))
		res := make(chan *InternalMessage)

		go func() {
			for val := range rawInput {
				bs := &InternalMessage{}
				err := json.Unmarshal(val, bs)
				if err == nil {
					res <- bs
				}
			}
		}()

		internalMessageInputChannels[ID] = res
	}
	return internalMessageInputChannels[ID]
}

func GetInternalMessageOutputChannel(ID party.ID) chan<- *InternalMessage {
	if internalMessageOutputChannels[ID] == nil {
		rawOutput := messaging.GetOutputChannel(messaging.InternalMessagesChannel + ":" + string(ID))
		res := make(chan *InternalMessage)

		go func() {
			for bs := range res {
				val, err := json.Marshal(bs)
				if err == nil {
					rawOutput <- val
				}
			}
		}()

		internalMessageOutputChannels[ID] = res
	}
	return internalMessageOutputChannels[ID]
}
