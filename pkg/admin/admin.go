package admin

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/j4y_funabashi/inari-micropub/pkg/app"
	"github.com/sirupsen/logrus"
)

type Server struct {
	App    app.Server
	logger *logrus.Logger
}

func NewServer(
	a app.Server,
	logger *logrus.Logger,
) Server {
	return Server{
		App:    a,
		logger: logger,
	}
}

func (s Server) Routes(router *mux.Router) {
	router.HandleFunc("/admin/media", s.handleMedia())
}

func (s Server) handleMedia() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		year := r.URL.Query().Get("year")
		month := r.URL.Query().Get("month")
		day := r.URL.Query().Get("day")

		media := s.App.ShowMedia(year, month, day)

		s.logger.WithField("media", media).Info("SHOW ME MEDIA!")
	}
}
