package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"mpc_poc/service"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/mux"
	"github.com/koteld/multi-party-sig/pkg/party"
)

var ids party.IDSlice

type Parameters struct {
	Threshold int    `json:"threshold"`
	Message   string `json:"message"`
	Amount    string `json:"amount"`
	To        string `json:"to"`
}

func GenerateKeys(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var parameters Parameters
	_ = json.NewDecoder(r.Body).Decode(&parameters)
	res := service.GenerateKeys(ids, parameters.Threshold)
	_ = json.NewEncoder(w).Encode(res)
}

func RefreshKeys(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var parameters Parameters
	_ = json.NewDecoder(r.Body).Decode(&parameters)

	service.RefreshKeys(ids, parameters.Threshold)
	_ = json.NewEncoder(w).Encode("Distributed keys were refreshed successfully")
}

func Sign(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var parameters Parameters
	_ = json.NewDecoder(r.Body).Decode(&parameters)
	res := service.Sign(ids, parameters.Threshold, parameters.Message)
	_ = json.NewEncoder(w).Encode(res)
}

func PreSign(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var parameters Parameters
	_ = json.NewDecoder(r.Body).Decode(&parameters)
	service.PreSign(ids)
	_ = json.NewEncoder(w).Encode("Pre-signature was created successfully")
}

func SignOnline(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var parameters Parameters
	_ = json.NewDecoder(r.Body).Decode(&parameters)
	messageHash := crypto.Keccak256Hash([]byte(parameters.Message))
	res := service.SignOnline(ids, messageHash)
	_ = json.NewEncoder(w).Encode(res)
}

func SendEth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var parameters Parameters
	_ = json.NewDecoder(r.Body).Decode(&parameters)
	txHash := service.SendEth(ids)
	_ = json.NewEncoder(w).Encode(txHash)
}

func GetOnline(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	online := service.GetOnline(ids)
	_ = json.NewEncoder(w).Encode(online)
}

func GetConfigs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	configs := service.GetConfigs(ids)
	_ = json.NewEncoder(w).Encode(configs)
}

func initializeRouter() {
	r := mux.NewRouter()

	r.HandleFunc("/keys/generate", GenerateKeys).Methods("POST")
	r.HandleFunc("/keys/refresh", RefreshKeys).Methods("POST")
	r.HandleFunc("/sign", Sign).Methods("POST")
	r.HandleFunc("/presign", PreSign).Methods("POST")
	r.HandleFunc("/signonline", SignOnline).Methods("POST")
	r.HandleFunc("/sendeth", SendEth).Methods("POST")

	r.HandleFunc("/online", GetOnline).Methods("GET")
	r.HandleFunc("/configs", GetConfigs).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", r))
}

func main() {
	idsArray := make([]party.ID, 0)
	for _, id := range os.Args[1:] {
		idsArray = append(idsArray, party.ID(id))
	}
	ids = party.NewIDSlice(idsArray)

	initializeRouter()
}
