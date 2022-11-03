package service

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"time"

	"mpc_poc/models"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/koteld/multi-party-sig/pkg/party"
	"github.com/lithammer/shortuuid"
)

var address common.Address
var publicKeyBytes []byte

func genShortUUID() string {
	return shortuuid.New()
}

func GenerateKeys(ids party.IDSlice, threshold int) models.ConfigMessage {
	if threshold == 0 {
		threshold = 1
	}
	results := make(map[party.ID][]byte, ids.Len())
	sessionID := genShortUUID()

	logMessages := models.GetLogMessageOutputChannel()
	logMessage := models.LogMessage{
		Protocol:    models.DKG,
		Participant: "initiator",
		Message:     "started protocol initialization",
		SessionID:   sessionID,
		Timestamp:   strconv.FormatInt(time.Now().Unix(), 10),
	}
	logMessages <- &logMessage

	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(id party.ID) {
			defer wg.Done()
			protocolMessages := models.GetProtocolMessageOutputChannel(id)
			protocolMessage := models.ProtocolMessage{
				Protocol:  models.DKG,
				IDs:       ids,
				Threshold: threshold,
				SessionID: []byte(sessionID),
			}
			protocolMessages <- &protocolMessage
			sessionMessagesChannel := models.GetSessionMessageInputChannel(sessionID, id)
			result := <-sessionMessagesChannel
			results[id], _ = b64.StdEncoding.DecodeString(result.Result.(string))
		}(id)
	}
	wg.Wait()

	publicKeyBytes = results[ids[0]]
	address = common.BytesToAddress(crypto.Keccak256(publicKeyBytes[1:])[12:])

	logMessage = models.LogMessage{
		Protocol:    models.DKG,
		Participant: "initiator",
		Message:     "protocol successfully completed",
		SessionID:   sessionID,
		Timestamp:   strconv.FormatInt(time.Now().Unix(), 10),
	}
	logMessages <- &logMessage

	return models.ConfigMessage{
		Address:   address.String(),
		IDs:       ids,
		SessionID: sessionID,
	}
}

func RefreshKeys(ids party.IDSlice, threshold int, address string) models.ConfigMessage {
	if threshold == 0 {
		threshold = 1
	}
	sessionID := genShortUUID()

	logMessages := models.GetLogMessageOutputChannel()
	logMessage := models.LogMessage{
		Protocol:    models.DKF,
		Participant: "initiator",
		Message:     "started protocol initialization",
		SessionID:   sessionID,
		Timestamp:   strconv.FormatInt(time.Now().Unix(), 10),
	}
	logMessages <- &logMessage

	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(id party.ID) {
			defer wg.Done()
			protocolMessages := models.GetProtocolMessageOutputChannel(id)
			protocolMessage := models.ProtocolMessage{
				Protocol:  models.DKF,
				IDs:       ids,
				Threshold: threshold,
				SessionID: []byte(sessionID),
				Address:   address,
			}
			protocolMessages <- &protocolMessage
			sessionMessagesChannel := models.GetSessionMessageInputChannel(sessionID, id)
			_ = <-sessionMessagesChannel
		}(id)
	}
	wg.Wait()

	logMessage = models.LogMessage{
		Protocol:    models.DKF,
		Participant: "initiator",
		Message:     "protocol successfully completed",
		SessionID:   sessionID,
		Timestamp:   strconv.FormatInt(time.Now().Unix(), 10),
	}
	logMessages <- &logMessage

	return models.ConfigMessage{
		Address:   address,
		IDs:       ids,
		SessionID: sessionID,
	}
}

func Sign(ids party.IDSlice, threshold int, messageHash common.Hash, address string) []byte {
	if threshold == 0 {
		threshold = 1
	}
	results := make(map[party.ID][]byte, ids.Len())
	sessionID := genShortUUID()

	logMessages := models.GetLogMessageOutputChannel()
	logMessage := models.LogMessage{
		Protocol:    models.Sign,
		Participant: "initiator",
		Message:     "started protocol initialization",
		SessionID:   sessionID,
		Timestamp:   strconv.FormatInt(time.Now().Unix(), 10),
	}
	logMessages <- &logMessage

	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(id party.ID) {
			defer wg.Done()
			protocolMessages := models.GetProtocolMessageOutputChannel(id)
			protocolMessage := models.ProtocolMessage{
				Protocol:    models.Sign,
				IDs:         ids,
				Threshold:   threshold,
				MessageHash: messageHash.Bytes(),
				SessionID:   []byte(sessionID),
				Address:     address,
			}
			protocolMessages <- &protocolMessage
			sessionMessagesChannel := models.GetSessionMessageInputChannel(sessionID, id)
			result := <-sessionMessagesChannel
			results[id], _ = b64.StdEncoding.DecodeString(result.Result.(string))
		}(id)
	}
	wg.Wait()

	logMessage = models.LogMessage{
		Protocol:    models.Sign,
		Participant: "initiator",
		Message:     "protocol successfully completed",
		SessionID:   sessionID,
		Timestamp:   strconv.FormatInt(time.Now().Unix(), 10),
	}
	logMessages <- &logMessage

	return results[ids[0]]
}

func PreSign(ids party.IDSlice, address string) {
	sessionID := genShortUUID()

	logMessages := models.GetLogMessageOutputChannel()
	logMessage := models.LogMessage{
		Protocol:    models.PreSign,
		Participant: "initiator",
		Message:     "started protocol initialization",
		SessionID:   sessionID,
		Timestamp:   strconv.FormatInt(time.Now().Unix(), 10),
	}
	logMessages <- &logMessage

	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(id party.ID) {
			defer wg.Done()
			protocolMessages := models.GetProtocolMessageOutputChannel(id)
			protocolMessage := models.ProtocolMessage{
				Protocol:  models.PreSign,
				IDs:       ids,
				SessionID: []byte(sessionID),
				Address:   address,
			}
			protocolMessages <- &protocolMessage
			sessionMessagesChannel := models.GetSessionMessageInputChannel(sessionID, id)
			result := <-sessionMessagesChannel
			fmt.Println(result)
		}(id)
	}
	wg.Wait()

	logMessage = models.LogMessage{
		Protocol:    models.PreSign,
		Participant: "initiator",
		Message:     "protocol successfully completed",
		SessionID:   sessionID,
		Timestamp:   strconv.FormatInt(time.Now().Unix(), 10),
	}
	logMessages <- &logMessage
}

func SignOnline(ids party.IDSlice, messageHash common.Hash, address string) []byte {
	results := make(map[party.ID][]byte, ids.Len())
	sessionID := genShortUUID()

	logMessages := models.GetLogMessageOutputChannel()
	logMessage := models.LogMessage{
		Protocol:    models.SignOnline,
		Participant: "initiator",
		Message:     "started protocol initialization",
		SessionID:   sessionID,
		Timestamp:   strconv.FormatInt(time.Now().Unix(), 10),
	}
	logMessages <- &logMessage

	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(id party.ID) {
			defer wg.Done()
			protocolMessages := models.GetProtocolMessageOutputChannel(id)
			protocolMessage := models.ProtocolMessage{
				Protocol:    models.SignOnline,
				IDs:         ids,
				MessageHash: messageHash.Bytes(),
				SessionID:   []byte(sessionID),
				Address:     address,
			}
			protocolMessages <- &protocolMessage
			sessionMessagesChannel := models.GetSessionMessageInputChannel(sessionID, id)
			result := <-sessionMessagesChannel
			results[id], _ = b64.StdEncoding.DecodeString(result.Result.(string))
		}(id)
	}
	wg.Wait()

	logMessage = models.LogMessage{
		Protocol:    models.SignOnline,
		Participant: "initiator",
		Message:     "protocol successfully completed",
		SessionID:   sessionID,
		Timestamp:   strconv.FormatInt(time.Now().Unix(), 10),
	}
	logMessages <- &logMessage

	return results[ids[0]]
}

func SendEth(ids party.IDSlice, threshold int, from string, to string, amount string, online bool) (string, error) {
	client, err := ethclient.Dial("https://goerli.infura.io/v3/4684d4b1567d4b78a9be3356bd3399b9")
	if err != nil {
		return "", err
	}

	fromAddress := common.HexToAddress(from)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", err
	}

	a, err := strconv.Atoi(amount)
	if err != nil {
		return "", err
	}
	value := big.NewInt(int64(a))
	tipCap, _ := client.SuggestGasTipCap(context.Background())
	feeCap, _ := client.SuggestGasPrice(context.Background())
	gasLimit := uint64(21000)
	if err != nil {
		return "", err
	}

	toAddress := common.HexToAddress(to)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return "", err
	}

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		GasTipCap: tipCap,
		GasFeeCap: feeCap,
		Nonce:     nonce,
		To:        &toAddress,
		Value:     value,
		Gas:       gasLimit,
		Data:      nil,
	})

	signer := types.NewLondonSigner(chainID)

	txHash := signer.Hash(tx)

	var sig []byte
	if online == true {
		sig = SignOnline(ids, txHash, from)
	} else {
		sig = Sign(ids, threshold, txHash, from)
	}

	signedTx, _ := tx.WithSignature(signer, sig)

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	hash := tx.Hash().Hex()
	return hash, nil
}

func GetOnline(ids party.IDSlice) map[party.ID]bool {
	results := make(map[party.ID]bool, ids.Len())
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(id party.ID) {
			defer wg.Done()
			infoRequestChannel := models.GetInfoRequestMessageOutputChannel(id)
			infoRequestMessage := models.InfoRequestMessage{
				Info: models.Online,
			}
			infoRequestChannel <- &infoRequestMessage

			infoResponseChannel := models.GetInfoResponseMessageInputChannel(id)
			result := <-infoResponseChannel
			results[id] = result.Online
		}(id)
	}
	wg.Wait()

	return results
}

func GetConfigs(ids party.IDSlice) []models.ConfigMessage {
	results := make(map[party.ID][]models.ConfigMessage, ids.Len())
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(id party.ID) {
			defer wg.Done()
			infoRequestChannel := models.GetInfoRequestMessageOutputChannel(id)
			infoRequestMessage := models.InfoRequestMessage{
				Info: models.Configs,
			}
			infoRequestChannel <- &infoRequestMessage

			infoResponseChannel := models.GetInfoResponseMessageInputChannel(id)
			result := <-infoResponseChannel
			results[id] = result.Configs
		}(id)
	}
	wg.Wait()

	configs := make(map[string]map[string]models.ConfigMessage)
	checkup := make(map[string]map[string]map[string]bool)

	for id, config := range results {
		for _, configMessage := range config {
			if configs[configMessage.Address] == nil {
				configs[configMessage.Address] = make(map[string]models.ConfigMessage)
				configs[configMessage.Address][configMessage.SessionID] = configMessage
			}

			if checkup[configMessage.Address] == nil {
				checkup[configMessage.Address] = make(map[string]map[string]bool)
			}
			if checkup[configMessage.Address][configMessage.SessionID] == nil {
				checkup[configMessage.Address][configMessage.SessionID] = make(map[string]bool)
			}
			if checkup[configMessage.Address][configMessage.SessionID][string(id)] == false {
				checkup[configMessage.Address][configMessage.SessionID][string(id)] = true
			}
		}
	}

	result := make([]models.ConfigMessage, 0)

	for address, c := range configs {
		for sessionId, config := range c {
			valid := true
			for _, id := range config.IDs {
				if checkup[address][sessionId][string(id)] == false {
					valid = false
					break
				}
			}
			if valid == true {
				result = append(result, config)
			}
		}
	}

	return result
}
