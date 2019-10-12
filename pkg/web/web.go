// Package web routes and handles web requests
package web

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/j4y_funabashi/inari-micropub/pkg/app"
	"github.com/j4y_funabashi/inari-micropub/pkg/view"
	"github.com/sirupsen/logrus"
)

type Server struct {
	App       app.Server
	logger    *logrus.Logger
	presenter view.Presenter
}

func NewServer(
	a app.Server,
	logger *logrus.Logger,
	p view.Presenter,
) Server {
	return Server{
		App:       a,
		logger:    logger,
		presenter: p,
	}
}

func (s Server) Routes(router *mux.Router) {
	router.HandleFunc("/admin/media", s.handleMediaGallery())
}

func (s Server) handleMediaGallery() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		year := r.URL.Query().Get("year")
		month := r.URL.Query().Get("month")
		day := r.URL.Query().Get("day")

		media := s.App.ShowMedia(year, month, day)
		s.logger.WithField("media", media).Info("SHOW ME MEDIA!")
		viewModel := s.presenter.ParseMediaGallery(media)
		err := renderMediaGallery(viewModel, w)
		if err != nil {
			s.logger.WithError(err).Error("failed to render media gallery")
		}

	}
}
