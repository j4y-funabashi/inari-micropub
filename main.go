package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/j4y_funabashi/inari-micropub/pkg/eventlog"
	"github.com/j4y_funabashi/inari-micropub/pkg/indieauth"
	"github.com/j4y_funabashi/inari-micropub/pkg/micropub"
	"github.com/j4y_funabashi/inari-micropub/pkg/s3"
	"github.com/sirupsen/logrus"
)

func main() {

	// config
	port := "80"
	tokenEndpoint := os.Getenv("TOKEN_ENDPOINT")
	mediaURL := os.Getenv("MEDIA_ENDPOINT")
	s3Endpoint := os.Getenv("S3_ENDPOINT")
	S3KeyPrefix := os.Getenv("S3_EVENTS_KEY")
	S3Bucket := os.Getenv("S3_EVENTS_BUCKET")

	// deps
	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}
	router := mux.NewRouter()

	eventSaver := s3.NewSaver(
		logger,
		s3Endpoint,
		S3KeyPrefix,
		S3Bucket,
	)

	// replay
	s3Client, err := s3.NewClient()
	if err != nil {
		logger.WithError(err).Error("failed to connect to s3")
		return
	}
	eventListing, err := eventlog.NewS3EventList(
		S3Bucket,
		S3KeyPrefix,
		s3Client,
		logger.WithField("pkg", "eventlog"),
	)
	if err != nil {
		logger.WithError(err).Error("failed to connect to event store")
	}
	postList, err := eventlog.Replay(
		eventListing,
		logger.WithField("pkg", "replay"),
	)
	if err != nil {
		logger.WithError(err).Error("failed to replay events")
	}

	micropubServer := micropub.NewServer(
		mediaURL,
		tokenEndpoint,
		logger,
		eventSaver,
		indieauth.VerifyAccessToken,
		&postList,
	)
	micropubServer.Routes(router)

	logger.Info("micropub server running on port " + port)
	logger.Fatal(http.ListenAndServe(":"+port, router))
}
