package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/j4y_funabashi/inari-micropub/pkg/db"
	"github.com/j4y_funabashi/inari-micropub/pkg/eventlog"
	"github.com/j4y_funabashi/inari-micropub/pkg/indieauth"
	"github.com/j4y_funabashi/inari-micropub/pkg/micropub"
	"github.com/j4y_funabashi/inari-micropub/pkg/s3"
	"github.com/sirupsen/logrus"
)

func main() {

	// config
	port := os.Getenv("PORT")
	tokenEndpoint := os.Getenv("TOKEN_ENDPOINT")
	mediaURL := os.Getenv("MEDIA_ENDPOINT")
	s3Endpoint := os.Getenv("S3_ENDPOINT")
	S3KeyPrefix := os.Getenv("S3_EVENTS_KEY")
	S3Bucket := os.Getenv("S3_EVENTS_BUCKET")
	mediaBucket := os.Getenv("S3_MEDIA_BUCKET")

	// deps
	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}
	router := mux.NewRouter()

	s3Client, err := s3.NewClient(s3Endpoint)
	if err != nil {
		logger.WithError(err).Error("failed to connect to s3")
		return
	}

	sqlDB, err := db.OpenDB()
	if err != nil {
		logger.WithError(err).Error("failed to open DB")
		return
	}

	selecta := db.NewSelecta(sqlDB)

	eventLog := eventlog.NewEventLog(
		S3KeyPrefix,
		S3Bucket,
		s3Client,
		sqlDB,
		logger,
	)

	go eventLog.Replay()

	mediaServer := micropub.NewMediaServer(s3Client, mediaURL, mediaBucket)

	micropubServer := micropub.NewServer(
		mediaURL,
		tokenEndpoint,
		logger,
		indieauth.VerifyAccessToken,
		selecta,
		eventLog,
		mediaServer,
	)
	micropubServer.Routes(router)

	logger.Info("XX micropub server running on port " + port)
	logger.Fatal(http.ListenAndServe(":"+port, router))
}
