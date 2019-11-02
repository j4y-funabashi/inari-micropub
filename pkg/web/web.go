// Package web routes and handles web requests
package web

import (
	"context"
	"net/http"
	"strconv"

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

type contextKey string

var contextKeySessionID = contextKey("session-id")

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
	router.HandleFunc("/admin/composer", s.adminOnly(s.handleComposerSubmit())).Methods("POST")
	router.HandleFunc("/admin/composer/media", s.adminOnly(s.handleMediaGallery())).Methods("GET")
	router.HandleFunc("/admin/composer/media", s.adminOnly(s.handleAddMediaToComposer())).Methods("POST")
	router.HandleFunc("/admin/composer/media/detail", s.adminOnly(s.handleMediaDetail())).Methods("GET")
	router.HandleFunc("/admin/composer/location", s.adminOnly(s.handleLocationSearch())).Methods("GET")
	router.HandleFunc("/admin/composer/location", s.adminOnly(s.handleAddLocationToComposer())).Methods("POST")
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
		sess, ok := r.Context().Value(contextKeySessionID).(app.SessionData)
		if !ok {
			s.logger.Error("failed fetch session from context")
		}
		viewModel := s.presenter.ParseComposer(sess)
		err := renderComposerForm(viewModel, w)
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
		w.Header().Set("Set-Cookie", authResponse.Session.CookieValue(1209600))
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
		sess, err := s.App.FetchSession(cookie.Value)

		if err != nil || sess.Token != cookie.Value {
			w.Header().Set("Location", "/login")
			w.Header().Set("Set-Cookie", sess.CookieValue(0))
			w.WriteHeader(http.StatusSeeOther)
			return
		}
		s.logger.
			WithField("session_id", cookie.Value).
			WithField("existing_session", sess).
			Info("found session")

		ctx := context.WithValue(r.Context(), contextKeySessionID, sess)
		h(w, r.WithContext(ctx))
	}
}

func (s Server) handleAddLocationToComposer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		lat, _ := strconv.ParseFloat(r.FormValue("lat"), 64)
		lng, _ := strconv.ParseFloat(r.FormValue("lng"), 64)
		loc := app.Location{
			Lat:      lat,
			Lng:      lng,
			Name:     r.FormValue("name"),
			Locality: r.FormValue("locality"),
			Region:   r.FormValue("region"),
			Country:  r.FormValue("country"),
		}

		// TODO fetch session from context
		sess, ok := r.Context().Value(contextKeySessionID).(app.SessionData)
		if !ok {
			s.logger.Error("failed fetch session from context")
		}

		sess.Location = loc

		s.logger.WithField("sess", sess).Info("added location to session")
		err := s.App.SaveSession(sess)
		if err != nil {
			s.logger.WithError(err).Error("failed save session")
		}
		w.Header().Set("Location", "/admin/composer")
		w.WriteHeader(http.StatusSeeOther)
	}
}

func (s Server) handleComposerSubmit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// TODO fetch session from context
		sess, ok := r.Context().Value(contextKeySessionID).(app.SessionData)
		if !ok {
			s.logger.Error("failed fetch session from context")
			return
		}

		sess.Content = r.FormValue("content")
		err := s.App.CreatePost(sess)
		if err != nil {
			s.logger.WithError(err).Error("failed to create post")
			return
		}
		sess = sess.Reset()

		s.logger.WithField("sess", sess).Info("session")
		err = s.App.SaveSession(sess)
		if err != nil {
			s.logger.WithError(err).Error("failed save session")
		}

		s.logger.Info("created post!")
		w.Header().Set("Location", "/admin/composer")
		w.WriteHeader(http.StatusSeeOther)
	}
}

func (s Server) handleAddMediaToComposer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mediaURL := r.FormValue("media_url")
		media := s.App.ShowMediaDetail(mediaURL)

		// TODO fetch session from context
		sess, ok := r.Context().Value(contextKeySessionID).(app.SessionData)
		if !ok {
			s.logger.Error("failed fetch session from context")
		}
		sess.Media = append(sess.Media, media.Media)
		sess.Published = media.Media.DateTime

		// TODO if media has lat/lng then search for suggested Locations
		if media.Media.Lat != 0 && media.Media.Lng != 0 {
			locations := s.App.SearchLocationsByLatLng(media.Media.Lat, media.Media.Lng)
			sess.AddSuggestedLocations(locations)
			venues := s.App.SearchVenues(media.Media.Lat, media.Media.Lng)
			sess.AddSuggestedLocations(venues)
		}

		s.logger.WithField("sess", sess).Info("session")
		err := s.App.SaveSession(sess)
		if err != nil {
			s.logger.WithError(err).Error("failed save session")
		}
		w.Header().Set("Location", "/admin/composer")
		w.WriteHeader(http.StatusSeeOther)
	}
}

func (s Server) handleLocationSearch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		locationQuery := r.URL.Query().Get("location")
		locations := s.App.SearchLocations(locationQuery)
		viewModel := s.presenter.ParseLocationSearch(locationQuery, locations)
		err := renderLocationSearch(viewModel, w)
		if err != nil {
			s.logger.WithError(err).Error("failed to render location search")
		}

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
