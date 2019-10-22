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
	router.HandleFunc("/login", s.handleLoginForm()).Methods("GET")
	router.HandleFunc("/login", s.handleLogin()).Methods("POST")
	router.HandleFunc("/admin/composer", s.adminOnly(s.handleComposerForm())).Methods("GET")
	router.HandleFunc("/admin/composer/media", s.adminOnly(s.handleMediaGallery())).Methods("GET")
	router.HandleFunc("/admin/composer/media/detail", s.adminOnly(s.handleMediaDetail())).Methods("GET")
}

func (s Server) handleMediaDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mediaURL := r.URL.Query().Get("url")
		mediaDetail := s.App.ShowMediaDetail(mediaURL)
		viewModel := s.presenter.ParseMediaDetail(mediaDetail)
		err := renderMediaDetail(viewModel, w)
		if err != nil {
			s.logger.WithError(err).Error("failed to render composer")
		}
	}
}

func (s Server) handleComposerForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := renderComposerForm(w)
		if err != nil {
			s.logger.WithError(err).Error("failed to render composer")
		}
	}
}

func (s Server) handleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enteredPassword := r.FormValue("password")
		authResponse, err := s.App.Auth(enteredPassword)
		if err != nil {
			s.logger.WithError(err).Error("failed to auth")
			w.Header().Set("Location", "/login")
			w.WriteHeader(http.StatusSeeOther)
			return
		}

		w.Header().Set("Location", "/admin/composer")
		w.Header().Set("Set-Cookie", authResponse.Session.CookieValue())
		w.WriteHeader(http.StatusSeeOther)
		return
	}
}

func (s Server) handleLoginForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := renderLoginForm(w)
		if err != nil {
			s.logger.WithError(err).Error("failed to render media gallery")
		}
	}
}

func (s Server) adminOnly(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			w.Header().Set("Location", "/login")
			w.WriteHeader(http.StatusSeeOther)
			return
		}
		s.logger.WithField("session_id", cookie.Value)
		h(w, r)
	}
}

func (s Server) handleMediaGallery() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		year := r.URL.Query().Get("year")
		month := r.URL.Query().Get("month")
		day := r.URL.Query().Get("day")

		media := s.App.ShowMediaGallery(year, month, day)
		viewModel := s.presenter.ParseMediaGallery(media)
		err := renderMediaGallery(viewModel, w)
		if err != nil {
			s.logger.WithError(err).Error("failed to render media gallery")
		}

	}
}
