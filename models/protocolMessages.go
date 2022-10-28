package models

import (
	"encoding/json"

	"mpc_poc/messaging"

	"github.com/koteld/multi-party-sig/pkg/party"
)

type Protocol string

const (
	DKG        Protocol = "protocol/dkg"
	DKF        Protocol = "protocol/dkf"
	Sign       Protocol = "protocol/sign"
	PreSign    Protocol = "protocol/presign"
	SignOnline Protocol = "protocol/signonline"
)

type (
	ProtocolMessage struct {
		Protocol     Protocol           `json:"protocol"`
		IDs          party.IDSlice      `json:"ids"`
		Threshold    int                `json:"threshold"`
		SessionID    []byte             `json:"sessionID"`
		MessageHash  []byte             `json:"messageHash"`
		PreSignature ecdsa.PreSignature `json:"preSignature"`
	}
)

var protocolMessageInputChannels = make(map[party.ID]<-chan *ProtocolMessage)
var protocolMessageOutputChannels = make(map[party.ID]chan<- *ProtocolMessage)

func GetProtocolMessageInputChannel(ID party.ID) <-chan *ProtocolMessage {
	if protocolMessageInputChannels[ID] == nil {
		rawInput := messaging.GetInputChannel(messaging.ProtocolMessagesChannel + ":" + string(ID))
		res := make(chan *ProtocolMessage)

		go func() {
			for val := range rawInput {
				bs := &ProtocolMessage{}
				err := json.Unmarshal(val, bs)
				if err == nil {
					res <- bs
				}
			}
		}()

		protocolMessageInputChannels[ID] = res
	}
	return protocolMessageInputChannels[ID]
}

func GetProtocolMessageOutputChannel(ID party.ID) chan<- *ProtocolMessage {
	if protocolMessageOutputChannels[ID] == nil {
		rawOutput := messaging.GetOutputChannel(messaging.ProtocolMessagesChannel + ":" + string(ID))
		res := make(chan *ProtocolMessage)

		go func() {
			for bs := range res {
				val, err := json.Marshal(bs)
				if err == nil {
					rawOutput <- val
				}
			}
		}()

		protocolMessageOutputChannels[ID] = res
	}
	return protocolMessageOutputChannels[ID]
}
