package service

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"fmt"
	"log"
	"math/big"
	"sync"

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

func GenerateKeys(ids party.IDSlice, threshold int) string {
	results := make(map[party.ID][]byte, ids.Len())
	sessionID := genShortUUID()
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

	return address.String()
}

func RefreshKeys(ids party.IDSlice, threshold int) {
	sessionID := genShortUUID()
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
			}
			protocolMessages <- &protocolMessage
			sessionMessagesChannel := models.GetSessionMessageInputChannel(sessionID, id)
			_ = <-sessionMessagesChannel
		}(id)
	}
	wg.Wait()
}

func Sign(ids party.IDSlice, threshold int, message string) []byte {
	results := make(map[party.ID][]byte, ids.Len())
	messageHash := crypto.Keccak256Hash([]byte(message))
	sessionID := genShortUUID()
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
			}
			protocolMessages <- &protocolMessage
			sessionMessagesChannel := models.GetSessionMessageInputChannel(sessionID, id)
			result := <-sessionMessagesChannel
			results[id], _ = b64.StdEncoding.DecodeString(result.Result.(string))
		}(id)
	}
	wg.Wait()

	return results[ids[0]]
}

func PreSign(ids party.IDSlice) {
	sessionID := genShortUUID()
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
			}
			protocolMessages <- &protocolMessage
			sessionMessagesChannel := models.GetSessionMessageInputChannel(sessionID, id)
			result := <-sessionMessagesChannel
			fmt.Println(result)
		}(id)
	}
	wg.Wait()
}

func SignOnline(ids party.IDSlice, messageHash common.Hash) []byte {
	results := make(map[party.ID][]byte, ids.Len())
	sessionID := genShortUUID()
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
			}
			protocolMessages <- &protocolMessage
			sessionMessagesChannel := models.GetSessionMessageInputChannel(sessionID, id)
			result := <-sessionMessagesChannel
			results[id], _ = b64.StdEncoding.DecodeString(result.Result.(string))
		}(id)
	}
	wg.Wait()

	return results[ids[0]]
}

func SendEth(ids party.IDSlice) string {
	client, err := ethclient.Dial("https://goerli.infura.io/v3/4684d4b1567d4b78a9be3356bd3399b9")
	if err != nil {
		log.Fatal(err)
	}

	nonce, err := client.PendingNonceAt(context.Background(), address)
	if err != nil {
		log.Fatal(err)
	}

	value := big.NewInt(1000000000000)
	tipCap, _ := client.SuggestGasTipCap(context.Background())
	feeCap, _ := client.SuggestGasPrice(context.Background())
	gasLimit := uint64(21000)
	if err != nil {
		log.Fatal(err)
	}

	toAddress := common.HexToAddress("0x4Ca389fAAd549aDd7124f2B215266cE162D964e7")

	chainID, err := client.NetworkID(context.Background())

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
	if err != nil {
		log.Fatal(err)
	}

	signer := types.NewLondonSigner(chainID)

	txHash := signer.Hash(tx)
	sig := SignOnline(ids, txHash)

	sigPublicKey, err := crypto.Ecrecover(txHash.Bytes(), sig)
	if err != nil {
		log.Fatal(err)
	}

	matches := bytes.Equal(sigPublicKey, publicKeyBytes)
	fmt.Println(matches) // true

	if err != nil {
		log.Fatal(err)
	}
	signedTx, _ := tx.WithSignature(signer, sig)

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	hash := tx.Hash().Hex()
	return hash
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

func GetConfigs(ids party.IDSlice) []string {
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

	addressesMap := make(map[string]bool)
	for _, config := range results {
		for _, configMessage := range config {
			addressesMap[configMessage.Address] = true
		}
	}

	addresses := make([]string, 0, len(addressesMap))
	for address := range addressesMap {
		addresses = append(addresses, address)
	}

	return addresses
}
