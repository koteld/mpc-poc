package main

import (
	"context"
	b64 "encoding/base64"
	"os"

	"mpc_poc/helper"
	"mpc_poc/models"
	"mpc_poc/session"
	mpcTypes "mpc_poc/types"

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
var configs = make(map[string]mpcTypes.Config)
var preSignatures = make(map[string]*ecdsa.PreSignature)

func saveConfigurationToFile(address string, sessionID []byte, config *cmp.Config) {
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
	bytes := make([]byte, len(sessionID)+len(configBytes))
	copy(bytes[0:22], sessionID)
	copy(bytes[22:], configBytes)

	_ = os.WriteFile(filepath, bytes, os.ModePerm)
}

func readConfigurationsFromFiles() {
	dir, _ := os.Getwd()
	filepath := dir + "/" + "participant-" + string(ID)
	files, _ := os.ReadDir(filepath)

	for _, file := range files {
		config := cmp.EmptyConfig(curve.Secp256k1{})
		fileData, _ := os.ReadFile(filepath + "/" + file.Name())

		sessionID := fileData[0:22]
		configBytes := fileData[22:]

		_ = config.UnmarshalBinary(configBytes)

		publicKeyBytes, _ := config.PublicPoint().MarshalBinaryEth()
		address := common.BytesToAddress(crypto.Keccak256(publicKeyBytes[1:])[12:])

		configs[address.String()] = mpcTypes.Config{
			Config:    config,
			SessionID: string(sessionID),
		}
	}
}

func startDKGProtocol(ids party.IDSlice, threshold int, sessionID []byte, pl *pool.Pool) {
	h, _ := protocol.NewMultiHandler(cmp.Keygen(curve.Secp256k1{}, ID, ids, threshold, pl), sessionID)
	session.Loop(ID, ids, h, string(sessionID), models.DKG)

	sessionMessageOutput := models.GetSessionMessageOutputChannel(string(sessionID), ID)
	r, err := h.Result()

	config := r.(*cmp.Config)
	publicKeyBytes, _ := config.PublicPoint().MarshalBinaryEth()
	address := common.BytesToAddress(crypto.Keccak256(publicKeyBytes[1:])[12:])

	saveConfigurationToFile(address.String(), sessionID, config)
	configs[address.String()] = mpcTypes.Config{
		Config:    config,
		SessionID: string(sessionID),
	}

	sessionMessage := models.SessionMessage{
		Result: b64.StdEncoding.EncodeToString(publicKeyBytes),
		Error:  err,
	}
	sessionMessageOutput <- &sessionMessage
}

func startDKFProtocol(address string, ids party.IDSlice, sessionID []byte, pl *pool.Pool) {
	h, _ := protocol.NewMultiHandler(cmp.Refresh(configs[address].Config, pl), sessionID)
	session.Loop(ID, ids, h, string(sessionID), models.DKF)

	sessionMessageOutput := models.GetSessionMessageOutputChannel(string(sessionID), ID)
	r, err := h.Result()

	config := r.(*cmp.Config)

	saveConfigurationToFile(address, sessionID, config)
	configs[address] = mpcTypes.Config{
		Config:    config,
		SessionID: string(sessionID),
	}

	sessionMessage := models.SessionMessage{
		Result: r,
		Error:  err,
	}
	sessionMessageOutput <- &sessionMessage
}

func startSignProtocol(address string, ids party.IDSlice, messageHash []byte, sessionID []byte, pl *pool.Pool) {
	h, _ := protocol.NewMultiHandler(cmp.Sign(configs[address].Config, ids, messageHash, pl), sessionID)
	session.Loop(ID, ids, h, string(sessionID), models.Sign)

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
	h, _ := protocol.NewMultiHandler(cmp.Presign(configs[address].Config, ids, pl), sessionID)
	session.Loop(ID, ids, h, string(sessionID), models.PreSign)

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
	h, _ := protocol.NewMultiHandler(cmp.PresignOnline(configs[address].Config, preSignatures[address], messageHash, pl), sessionID)
	session.Loop(ID, ids, h, string(sessionID), models.SignOnline)

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
			Address:   address,
			IDs:       config.Config.PartyIDs(),
			SessionID: config.SessionID,
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
