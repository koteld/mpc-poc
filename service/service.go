package service

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/koteld/multi-party-sig/pkg/party"
	"github.com/lithammer/shortuuid"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
)

func genShortUUID() string {
	return shortuuid.New()
}

func GenerateKeys(ids party.IDSlice, threshold int) {
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
			fmt.Println(result)
		}(id)
	}
	wg.Wait()
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
			result := <-sessionMessagesChannel
			fmt.Println(result)
		}(id)
	}
	wg.Wait()
}

func Sign(ids party.IDSlice, threshold int, message string) {
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
			fmt.Println(result)
		}(id)
	}
	wg.Wait()
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

func SignOnline(ids party.IDSlice, message string) {
	messageHash := crypto.Keccak256Hash([]byte(message))
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
			fmt.Println(result)
		}(id)
	}
	wg.Wait()
}
