package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"mpc_poc/broker"
	"mpc_poc/models"
	"mpc_poc/service"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/koteld/multi-party-sig/pkg/party"
)

var ids party.IDSlice

type Parameters struct {
	Threshold int    `json:"threshold"`
	Message   string `json:"message"`
	Address   string `json:"address"`
	To        string `json:"to"`
	Amount    string `json:"amount"`
	Online    bool   `json:"online"`
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

	res := service.RefreshKeys(ids, parameters.Threshold, parameters.Address)
	_ = json.NewEncoder(w).Encode(res)
}

func Sign(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var parameters Parameters
	_ = json.NewDecoder(r.Body).Decode(&parameters)
	messageHash := crypto.Keccak256Hash([]byte(parameters.Message))
	res := service.Sign(ids, parameters.Threshold, messageHash, parameters.Address)
	_ = json.NewEncoder(w).Encode(res)
}

func PreSign(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var parameters Parameters
	_ = json.NewDecoder(r.Body).Decode(&parameters)
	service.PreSign(ids, parameters.Address)
	_ = json.NewEncoder(w).Encode("Pre-signature was created successfully")
}

func SignOnline(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var parameters Parameters
	_ = json.NewDecoder(r.Body).Decode(&parameters)
	messageHash := crypto.Keccak256Hash([]byte(parameters.Message))
	res := service.SignOnline(ids, messageHash, parameters.Address)
	_ = json.NewEncoder(w).Encode(res)
}

func SendEth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var parameters Parameters
	_ = json.NewDecoder(r.Body).Decode(&parameters)
	txHash, err := service.SendEth(ids, parameters.Threshold, parameters.Address, parameters.To, parameters.Amount, parameters.Online)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(err)
	} else {
		_ = json.NewEncoder(w).Encode(txHash)
	}
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
	b := broker.NewServer()
	r := mux.NewRouter()

	logChannel := models.GetLogMessageInputChannel()
	go listenLogs(b, logChannel)

	r.HandleFunc("/keys/generate", GenerateKeys).Methods("POST")
	r.HandleFunc("/keys/refresh", RefreshKeys).Methods("POST")
	r.HandleFunc("/sign", Sign).Methods("POST")
	r.HandleFunc("/presign", PreSign).Methods("POST")
	r.HandleFunc("/signonline", SignOnline).Methods("POST")
	r.HandleFunc("/sendeth", SendEth).Methods("POST")

	r.HandleFunc("/online", GetOnline).Methods("GET")
	r.HandleFunc("/configs", GetConfigs).Methods("GET")

	r.HandleFunc("/sse", b.Stream).Methods("GET")

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	credentialsOk := handlers.AllowCredentials()

	log.Fatal(http.ListenAndServe(":8080", handlers.CORS(originsOk, headersOk, methodsOk, credentialsOk)(r)))
}

func listenLogs(b *broker.Broker, logChannel <-chan *models.LogMessage) {
	for {
		select {
		case logMessage := <-logChannel:
			j, _ := json.Marshal(logMessage)
			b.Notifier <- j
		}
	}
}

func main() {
	idsArray := make([]party.ID, 0)
	for _, id := range os.Args[1:] {
		idsArray = append(idsArray, party.ID(id))
	}
	ids = party.NewIDSlice(idsArray)

	initializeRouter()
}
