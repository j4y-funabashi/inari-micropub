package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/j4y_funabashi/inari-micropub/pkg/app"
	"github.com/j4y_funabashi/inari-micropub/pkg/db"
	"github.com/j4y_funabashi/inari-micropub/pkg/eventlog"
	"github.com/j4y_funabashi/inari-micropub/pkg/geocoder"
	"github.com/j4y_funabashi/inari-micropub/pkg/indieauth"
	"github.com/j4y_funabashi/inari-micropub/pkg/micropub"
	"github.com/j4y_funabashi/inari-micropub/pkg/s3"
	"github.com/j4y_funabashi/inari-micropub/pkg/session"
	"github.com/j4y_funabashi/inari-micropub/pkg/view"
	"github.com/j4y_funabashi/inari-micropub/pkg/web"
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
	geoAPIKey := os.Getenv("GEO_API_KEY")
	geoBaseURL := os.Getenv("GEO_API_URL")

	// deps
	logger := logrus.New()

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

	mediaServer := micropub.NewMediaServer(
		s3Client,
		mediaURL,
		mediaBucket,
	)

	micropubServer := micropub.NewServer(
		mediaURL,
		tokenEndpoint,
		logger,
		indieauth.VerifyAccessToken,
		selecta,
		eventLog,
		mediaServer,
	)
	sessStore := session.NewSessionStore(sqlDB)

	geo := geocoder.New(geoAPIKey, geoBaseURL, logger)

	inari := app.New(
		selecta,
		logger,
		sessStore,
		geo,
		eventLog,
	)
	presenter := view.NewPresenter()
	reqParser := web.NewParser()
	webServer := web.NewServer(
		inari,
		logger,
		presenter,
		reqParser,
	)

	// routing
	router := mux.NewRouter()
	router.StrictSlash(true)
	micropubServer.Routes(router.PathPrefix("/micropub").Subrouter())
	webServer.Routes(router.PathPrefix("/api").Subrouter())

	go eventLog.Replay()

	logger.Info("XX micropub server running on port " + port)
	logger.Fatal(http.ListenAndServe(":"+port, router))
}
