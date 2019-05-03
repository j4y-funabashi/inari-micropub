package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {

	// config
	port := "80"

	// deps
	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}
	router := mux.NewRouter()

	router.HandleFunc("/", handleTokens(logger))

	logger.Info("token mock server running on port " + port)
	logger.Fatal(http.ListenAndServe(":"+port, router))
}

func handleTokens(logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.
			WithField("Accept", r.Header.Get("Accept")).
			WithField("Authorization", r.Header.Get("Authorization")).
			Info("recieved request")

		if r.Header.Get("Authorization") == "Bearer test-valid-token" {
			jsonres := struct {
				Me       string `json:"me"`
				ClientID string `json:"client_id"`
				Scope    string `json:"scope"`
			}{
				Me:       "https://user.example.net/",
				ClientID: "https://app.example.com/",
				Scope:    "create update delete",
			}
			w.WriteHeader(200)
			json.NewEncoder(w).Encode(jsonres)
			return
		}
	}
}
