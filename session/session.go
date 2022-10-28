package session

import (
	"mpc_poc/models"

	"github.com/koteld/multi-party-sig/pkg/party"
	"github.com/koteld/multi-party-sig/pkg/protocol"
)

func SendMessage(msg *protocol.Message, ids party.IDSlice) {
	for _, id := range ids {
		if msg.IsFor(id) {
			internalMessageOutput := models.GetInternalMessageOutputChannel(id)
			internalMessage := models.InternalMessage{Message: *msg}
			internalMessageOutput <- &internalMessage
		}
	}
}

func Loop(id party.ID, ids party.IDSlice, h protocol.Handler) {
	internalMessageInput := models.GetInternalMessageInputChannel(id)

	for {
		select {
		// outgoing messages
		case msg, ok := <-h.Listen():
			if !ok {
				return
			}
			go SendMessage(msg, ids)
		// incoming messages
		case internalMessage := <-internalMessageInput:
			h.Accept(&internalMessage.Message)
		}
	}
}
