package models

import (
	"encoding/json"
	"sync"

	"mpc_poc/messaging"

	"github.com/taurusgroup/multi-party-sig/pkg/party"
)

type (
	SessionMessage struct {
		Result interface{} `json:"result"`
		Error  error       `json:"error"`
	}
)

var sessionMessageInputChannels = make(map[string]map[party.ID]<-chan *SessionMessage)
var sessionMessageOutputChannels = make(map[string]map[party.ID]chan<- *SessionMessage)

var mtx sync.Mutex

func GetSessionMessageInputChannel(SessionID string, ID party.ID) <-chan *SessionMessage {
	mtx.Lock()
	defer mtx.Unlock()
	if sessionMessageInputChannels[SessionID][ID] == nil {
		if sessionMessageInputChannels[SessionID] == nil {
			sessionMessageInputChannels[SessionID] = make(map[party.ID]<-chan *SessionMessage)
		}
		rawInput := messaging.GetInputChannel(messaging.SessionMessagesChannel + ":" + string(SessionID) + ":" + string(ID))
		res := make(chan *SessionMessage)

		go func() {
			for val := range rawInput {
				bs := &SessionMessage{}
				err := json.Unmarshal(val, bs)
				if err == nil {
					res <- bs
				}
			}
		}()

		sessionMessageInputChannels[SessionID][ID] = res
	}
	return sessionMessageInputChannels[SessionID][ID]
}

func GetSessionMessageOutputChannel(SessionID string, ID party.ID) chan<- *SessionMessage {
	mtx.Lock()
	defer mtx.Unlock()
	if sessionMessageOutputChannels[SessionID][ID] == nil {
		if sessionMessageOutputChannels[SessionID] == nil {
			sessionMessageOutputChannels[SessionID] = make(map[party.ID]chan<- *SessionMessage)
		}
		rawOutput := messaging.GetOutputChannel(messaging.SessionMessagesChannel + ":" + string(SessionID) + ":" + string(ID))
		res := make(chan *SessionMessage)

		go func() {
			for bs := range res {
				val, err := json.Marshal(bs)
				if err == nil {
					rawOutput <- val
				}
			}
		}()

		sessionMessageOutputChannels[SessionID][ID] = res
	}
	return sessionMessageOutputChannels[SessionID][ID]
}
