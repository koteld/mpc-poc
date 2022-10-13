package main

import (
	"context"
	"os"

	"mpc_poc/models"
	"mpc_poc/session"

	"github.com/taurusgroup/multi-party-sig/pkg/ecdsa"
	"github.com/taurusgroup/multi-party-sig/pkg/math/curve"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
	"github.com/taurusgroup/multi-party-sig/pkg/pool"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
	"github.com/taurusgroup/multi-party-sig/protocols/cmp"
)

var config *cmp.Config
var preSignature *ecdsa.PreSignature

var ID party.ID

func StartDKGProtocol(id party.ID, ids party.IDSlice, threshold int, sessionID []byte, pl *pool.Pool) {
	h, _ := protocol.NewMultiHandler(cmp.Keygen(curve.Secp256k1{}, id, ids, threshold, pl), sessionID)
	session.Loop(id, ids, h)

	sessionMessageOutput := models.GetSessionMessageOutputChannel(string(sessionID), id)
	r, err := h.Result()
	sessionMessage := models.SessionMessage{
		Result: r,
		Error:  err,
	}
	sessionMessageOutput <- &sessionMessage

	config = r.(*cmp.Config)
}

func StartDKFProtocol(id party.ID, ids party.IDSlice, sessionID []byte, pl *pool.Pool) {
	h, _ := protocol.NewMultiHandler(cmp.Refresh(config, pl), sessionID)
	session.Loop(id, ids, h)

	sessionMessageOutput := models.GetSessionMessageOutputChannel(string(sessionID), id)
	r, err := h.Result()
	sessionMessage := models.SessionMessage{
		Result: r,
		Error:  err,
	}
	sessionMessageOutput <- &sessionMessage

	config = r.(*cmp.Config)
}

func StartSignProtocol(id party.ID, ids party.IDSlice, messageHash []byte, sessionID []byte, pl *pool.Pool) {
	h, _ := protocol.NewMultiHandler(cmp.Sign(config, ids, messageHash, pl), sessionID)
	session.Loop(id, ids, h)

	sessionMessageOutput := models.GetSessionMessageOutputChannel(string(sessionID), id)
	r, err := h.Result()
	sessionMessage := models.SessionMessage{
		Result: r,
		Error:  err,
	}
	sessionMessageOutput <- &sessionMessage
}

func StartPreSignProtocol(id party.ID, ids party.IDSlice, sessionID []byte, pl *pool.Pool) {
	h, _ := protocol.NewMultiHandler(cmp.Presign(config, ids, pl), sessionID)
	session.Loop(id, ids, h)

	sessionMessageOutput := models.GetSessionMessageOutputChannel(string(sessionID), id)
	r, err := h.Result()
	sessionMessage := models.SessionMessage{
		Result: r,
		Error:  err,
	}
	sessionMessageOutput <- &sessionMessage

	preSignature = r.(*ecdsa.PreSignature)
}

func StartSignOnlineProtocol(id party.ID, ids party.IDSlice, messageHash []byte, sessionID []byte, pl *pool.Pool) {
	h, _ := protocol.NewMultiHandler(cmp.PresignOnline(config, preSignature, messageHash, pl), sessionID)
	session.Loop(id, ids, h)

	sessionMessageOutput := models.GetSessionMessageOutputChannel(string(sessionID), id)
	r, err := h.Result()
	sessionMessage := models.SessionMessage{
		Result: r,
		Error:  err,
	}
	sessionMessageOutput <- &sessionMessage
}

func StartProtocol(message *models.ProtocolMessage) {
	pl := pool.NewPool(0)
	defer pl.TearDown()

	switch message.Protocol {
	case models.DKG:
		StartDKGProtocol(ID, message.IDs, message.Threshold, message.SessionID, pl)
	case models.DKF:
		StartDKFProtocol(ID, message.IDs, message.SessionID, pl)
	case models.Sign:
		StartSignProtocol(ID, message.IDs, message.MessageHash, message.SessionID, pl)
	case models.PreSign:
		StartPreSignProtocol(ID, message.IDs, message.SessionID, pl)
	case models.SignOnline:
		StartSignOnlineProtocol(ID, message.IDs, message.MessageHash, message.SessionID, pl)
	}
}

func Activate(_ context.Context) {
	protocolMessageInput := models.GetProtocolMessageInputChannel(ID)

	for {
		select {
		case protocolMessage := <-protocolMessageInput:
			StartProtocol(protocolMessage)
		}
	}
}

func main() {
	ctx := context.Background()
	ID = party.ID(os.Args[1])
	Activate(ctx)
}
