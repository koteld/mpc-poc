package session

import (
	"strconv"
	"time"

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

func Loop(id party.ID, ids party.IDSlice, h protocol.Handler, sessionID string, protocol models.Protocol, ip string) {
	internalMessageInput := models.GetInternalMessageInputChannel(id)

	logMessages := models.GetLogMessageOutputChannel()
	logMessage := models.LogMessage{
		Protocol:    protocol,
		Participant: string(id),
		Message:     "protocol initialized",
		SessionID:   sessionID,
		Timestamp:   strconv.FormatInt(time.Now().Unix(), 10),
		IP:          ip,
	}
	logMessages <- &logMessage

	for {
		select {
		// outgoing messages
		case msg, ok := <-h.Listen():
			if !ok {
				logMessage = models.LogMessage{
					Protocol:    protocol,
					Participant: string(id),
					Message:     "all rounds completed",
					SessionID:   sessionID,
					Timestamp:   strconv.FormatInt(time.Now().Unix(), 10),
					IP:          ip,
				}
				logMessages <- &logMessage
				return
			}
			var to string
			if len(string(msg.To)) > 0 {
				to = string(msg.To)
			} else {
				to = "all"
			}
			logMessage = models.LogMessage{
				Protocol:    protocol,
				Participant: string(id),
				Message:     "sending message to: " + to,
				SessionID:   sessionID,
				Timestamp:   strconv.FormatInt(time.Now().Unix(), 10),
				Round:       uint16(msg.RoundNumber),
				IP:          ip,
			}
			logMessages <- &logMessage
			go SendMessage(msg, ids)
		// incoming messages
		case internalMessage := <-internalMessageInput:
			h.Accept(&internalMessage.Message)
			logMessage = models.LogMessage{
				Protocol:    protocol,
				Participant: string(id),
				Message:     "received message from: " + string(internalMessage.Message.From),
				SessionID:   sessionID,
				Timestamp:   strconv.FormatInt(time.Now().Unix(), 10),
				Round:       uint16(internalMessage.Message.RoundNumber),
				IP:          ip,
			}
			logMessages <- &logMessage
		}
	}
}
