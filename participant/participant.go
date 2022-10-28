package main

import (
	"context"
	b64 "encoding/base64"
	"os"

	"mpc_poc/helper"
	"mpc_poc/models"
	"mpc_poc/session"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/koteld/multi-party-sig/pkg/ecdsa"
	"github.com/koteld/multi-party-sig/pkg/math/curve"
	"github.com/koteld/multi-party-sig/pkg/party"
	"github.com/koteld/multi-party-sig/pkg/pool"
	"github.com/koteld/multi-party-sig/pkg/protocol"
	"github.com/koteld/multi-party-sig/protocols/cmp"
)

var ID party.ID
var configs = make(map[string]*cmp.Config)
var preSignatures = make(map[string]*ecdsa.PreSignature)

func saveConfigurationToFile(address string, config *cmp.Config) {
	wd, _ := os.Getwd()

	directory := wd + "/" + "participant-" + string(ID)
	if !helper.PathExists(directory) {
		_ = os.Mkdir(directory, os.ModePerm)
	}

	filepath := directory + "/" + address
	if helper.PathExists(filepath) {
		_ = os.Remove(filepath)
	}

	configBytes, _ := config.MarshalBinary()
	_ = os.WriteFile(filepath, configBytes, os.ModePerm)
}

func readConfigurationsFromFiles() {
	dir, _ := os.Getwd()
	filepath := dir + "/" + "participant-" + string(ID)
	files, _ := os.ReadDir(filepath)

	for _, file := range files {
		config := cmp.EmptyConfig(curve.Secp256k1{})
		fileData, _ := os.ReadFile(filepath + "/" + file.Name())
		_ = config.UnmarshalBinary(fileData)

		publicKeyBytes, _ := config.PublicPoint().MarshalBinaryEth()
		address := common.BytesToAddress(crypto.Keccak256(publicKeyBytes[1:])[12:])

		configs[address.String()] = config
	}
}

func startDKGProtocol(ids party.IDSlice, threshold int, sessionID []byte, pl *pool.Pool) {
	h, _ := protocol.NewMultiHandler(cmp.Keygen(curve.Secp256k1{}, ID, ids, threshold, pl), sessionID)
	session.Loop(ID, ids, h)

	sessionMessageOutput := models.GetSessionMessageOutputChannel(string(sessionID), ID)
	r, err := h.Result()

	config := r.(*cmp.Config)
	publicKeyBytes, _ := config.PublicPoint().MarshalBinaryEth()
	address := common.BytesToAddress(crypto.Keccak256(publicKeyBytes[1:])[12:])

	saveConfigurationToFile(address.String(), config)
	configs[address.String()] = config

	sessionMessage := models.SessionMessage{
		Result: b64.StdEncoding.EncodeToString(publicKeyBytes),
		Error:  err,
	}
	sessionMessageOutput <- &sessionMessage
}

func startDKFProtocol(address string, ids party.IDSlice, sessionID []byte, pl *pool.Pool) {
	h, _ := protocol.NewMultiHandler(cmp.Refresh(configs[address], pl), sessionID)
	session.Loop(ID, ids, h)

	sessionMessageOutput := models.GetSessionMessageOutputChannel(string(sessionID), ID)
	r, err := h.Result()

	config := r.(*cmp.Config)

	saveConfigurationToFile(address, config)
	configs[address] = config

	sessionMessage := models.SessionMessage{
		Result: r,
		Error:  err,
	}
	sessionMessageOutput <- &sessionMessage
}

func startSignProtocol(address string, ids party.IDSlice, messageHash []byte, sessionID []byte, pl *pool.Pool) {
	h, _ := protocol.NewMultiHandler(cmp.Sign(configs[address], ids, messageHash, pl), sessionID)
	session.Loop(ID, ids, h)

	sessionMessageOutput := models.GetSessionMessageOutputChannel(string(sessionID), ID)
	r, err := h.Result()
	signature := r.(*ecdsa.Signature)
	signatureCompact := signature.ToCompactEth()
	sessionMessage := models.SessionMessage{
		Result: b64.StdEncoding.EncodeToString(signatureCompact),
		Error:  err,
	}
	sessionMessageOutput <- &sessionMessage
}

func startPreSignProtocol(address string, ids party.IDSlice, sessionID []byte, pl *pool.Pool) {
	h, _ := protocol.NewMultiHandler(cmp.Presign(configs[address], ids, pl), sessionID)
	session.Loop(ID, ids, h)

	sessionMessageOutput := models.GetSessionMessageOutputChannel(string(sessionID), ID)
	r, err := h.Result()

	preSignatures[address] = r.(*ecdsa.PreSignature)

	sessionMessage := models.SessionMessage{
		Result: r,
		Error:  err,
	}
	sessionMessageOutput <- &sessionMessage
}

func startSignOnlineProtocol(address string, ids party.IDSlice, messageHash []byte, sessionID []byte, pl *pool.Pool) {
	h, _ := protocol.NewMultiHandler(cmp.PresignOnline(configs[address], preSignatures[address], messageHash, pl), sessionID)
	session.Loop(ID, ids, h)

	sessionMessageOutput := models.GetSessionMessageOutputChannel(string(sessionID), ID)
	r, err := h.Result()
	signature := r.(*ecdsa.Signature)
	signatureCompact := signature.ToCompactEth()
	sessionMessage := models.SessionMessage{
		Result: b64.StdEncoding.EncodeToString(signatureCompact),
		Error:  err,
	}
	sessionMessageOutput <- &sessionMessage
}

func startProtocol(message *models.ProtocolMessage) {
	pl := pool.NewPool(0)
	defer pl.TearDown()

	switch message.Protocol {
	case models.DKG:
		startDKGProtocol(message.IDs, message.Threshold, message.SessionID, pl)
	case models.DKF:
		startDKFProtocol(message.Address, message.IDs, message.SessionID, pl)
	case models.Sign:
		startSignProtocol(message.Address, message.IDs, message.MessageHash, message.SessionID, pl)
	case models.PreSign:
		startPreSignProtocol(message.Address, message.IDs, message.SessionID, pl)
	case models.SignOnline:
		startSignOnlineProtocol(message.Address, message.IDs, message.MessageHash, message.SessionID, pl)
	}
}

func getOnline() {
	infoMessageOutput := models.GetInfoResponseMessageOutputChannel(ID)
	infoMessage := models.InfoResponseMessage{
		Info:   models.Online,
		Online: true,
	}
	infoMessageOutput <- &infoMessage
}

func getConfigs() {
	infoMessageOutput := models.GetInfoResponseMessageOutputChannel(ID)
	configMessages := make([]models.ConfigMessage, 0, len(configs))
	for address, config := range configs {
		configMessage := models.ConfigMessage{
			Address: address,
			IDs:     config.PartyIDs(),
		}
		configMessages = append(configMessages, configMessage)
	}
	infoMessage := models.InfoResponseMessage{
		Info:    models.Online,
		Configs: configMessages,
	}
	infoMessageOutput <- &infoMessage
}

func getInfo(message *models.InfoRequestMessage) {
	pl := pool.NewPool(0)
	defer pl.TearDown()

	switch message.Info {
	case models.Online:
		getOnline()
	case models.Configs:
		getConfigs()
	}
}

func activate(_ context.Context) {
	readConfigurationsFromFiles()

	infoMessageInput := models.GetInfoRequestMessageInputChannel(ID)
	protocolMessageInput := models.GetProtocolMessageInputChannel(ID)

	for {
		select {
		case infoMessage := <-infoMessageInput:
			getInfo(infoMessage)
		case protocolMessage := <-protocolMessageInput:
			startProtocol(protocolMessage)
		}
	}
}

func main() {
	ctx := context.Background()
	ID = party.ID(os.Args[1])
	activate(ctx)
}
