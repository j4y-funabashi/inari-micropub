package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/j4y_funabashi/inari-micropub/pkg/micropub"
	"github.com/sirupsen/logrus"
)

func main() {

	// config
	port := "80"

	// deps
	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}
	router := mux.NewRouter()

	mediaURL := os.Getenv("MEDIA_ENDPOINT")
	micropubServer := micropub.NewServer(
		mediaURL,
		logger,
	)
	micropubServer.Routes(router)

	logger.Info("server running on port " + port)
	logger.Fatal(http.ListenAndServe(":"+port, router))
}
