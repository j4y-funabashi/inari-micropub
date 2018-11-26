package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/j4y_funabashi/inari-micropub/pkg/indieauth"
	"github.com/j4y_funabashi/inari-micropub/pkg/micropub"
	"github.com/j4y_funabashi/inari-micropub/pkg/s3"
	"github.com/sirupsen/logrus"
)

func main() {

	// config
	port := "80"

	// deps
	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}
	router := mux.NewRouter()

	// s3 saver
	s3Endpoint := os.Getenv("S3_ENDPOINT")
	S3KeyPrefix := os.Getenv("S3_EVENTS_KEY")
	S3Bucket := os.Getenv("S3_EVENTS_BUCKET")
	saveEvent := s3.NewSaver(
		logger,
		s3Endpoint,
		S3KeyPrefix,
		S3Bucket,
	)

	tokenEndpoint := os.Getenv("TOKEN_ENDPOINT")
	mediaURL := os.Getenv("MEDIA_ENDPOINT")
	micropubServer := micropub.NewServer(
		mediaURL,
		tokenEndpoint,
		logger,
		saveEvent,
		indieauth.VerifyAccessToken,
	)
	micropubServer.Routes(router)

	logger.Info("micropub server running on port " + port)
	logger.Fatal(http.ListenAndServe(":"+port, router))
}
