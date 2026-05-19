package main

import (
	"log"
	"maestro/internal/config"
	"net/http"

	"github.com/bytedance/sonic"
)

// var testIssue string = "MAE-1"
var ready bool

type Backend struct {
	Port  string
	ID    int
	Ready bool
}

func (api *Backend) ReadyEndpoint(w http.ResponseWriter, r *http.Request) {
	api.Ready = true
	resp := map[string]interface{}{
		"apiId": api.ID,
		"ready": api.Ready,
	}
	jsonData, err := sonic.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}
	w.Write(jsonData)
}

func main() {
	cfg, _ := config.LoadConfig()
	api := Backend{
		Port:  cfg.API.Port,
		ID:    1,
		Ready: ready,
	}

	router := http.NewServeMux()
	router.HandleFunc("/ready", api.ReadyEndpoint)

	server := &http.Server{
		Addr:    ":" + api.Port,
		Handler: router,
	}
	log.Printf("Server interno up : ID %v", api.ID)
	log.Fatal(server.ListenAndServe())
}
