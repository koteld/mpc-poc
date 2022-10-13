package main

import (
	"encoding/json"
	"log"
	"net/http"

	"mpc_poc/service"

	"github.com/gorilla/mux"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
)

type Parameters struct {
	IDs       party.IDSlice `json:"ids"`
	Threshold int           `json:"threshold"`
	Message   string        `json:"message"`
}

func GenerateKeys(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var parameters Parameters
	_ = json.NewDecoder(r.Body).Decode(&parameters)
	service.GenerateKeys(parameters.IDs, parameters.Threshold)
	_ = json.NewEncoder(w).Encode("Distributed keys were generated successfully")
}

func RefreshKeys(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var parameters Parameters
	_ = json.NewDecoder(r.Body).Decode(&parameters)

	service.RefreshKeys(parameters.IDs, parameters.Threshold)
	_ = json.NewEncoder(w).Encode("Distributed keys were refreshed successfully")
}

func Sign(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var parameters Parameters
	_ = json.NewDecoder(r.Body).Decode(&parameters)
	service.Sign(parameters.IDs, parameters.Threshold, parameters.Message)
	_ = json.NewEncoder(w).Encode("Message was signed successfully")
}

func PreSign(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var parameters Parameters
	_ = json.NewDecoder(r.Body).Decode(&parameters)
	service.PreSign(parameters.IDs)
	_ = json.NewEncoder(w).Encode("Pre-signature was created successfully")
}

func SignOnline(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var parameters Parameters
	_ = json.NewDecoder(r.Body).Decode(&parameters)
	service.SignOnline(parameters.IDs, parameters.Message)
	_ = json.NewEncoder(w).Encode("Message was signed online successfully")
}

func initializeRouter() {
	r := mux.NewRouter()

	r.HandleFunc("/keys/generate", GenerateKeys).Methods("POST")
	r.HandleFunc("/keys/refresh", RefreshKeys).Methods("POST")
	r.HandleFunc("/sign", Sign).Methods("POST")
	r.HandleFunc("/presign", PreSign).Methods("POST")
	r.HandleFunc("/signonline", SignOnline).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", r))
}

func main() {
	initializeRouter()
}
