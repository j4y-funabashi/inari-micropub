// Package web routes and handles web requests
package web

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
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
	parser    Parser
}

type contextKey string

var contextKeySessionID = contextKey("session-id")
var contextKeyAccessToken = contextKey("access-token")

func NewServer(
	a app.Server,
	logger *logrus.Logger,
	p view.Presenter,
	parser Parser,
) Server {
	return Server{
		App:       a,
		logger:    logger,
		presenter: p,
		parser:    parser,
	}
}

func (s Server) Routes(router *mux.Router) {
	router.HandleFunc("/", s.handleHomepage()).Methods("GET")
	router.HandleFunc("/micropub", s.withBearerToken(s.handleMicropubCommand())).Methods("POST")
	router.HandleFunc("/micropub", s.withBearerToken(s.handleMicropubQuery())).Methods("GET")
	router.HandleFunc("/login", s.handleLoginForm()).Methods("GET")
	router.HandleFunc("/login", s.handleLogin()).Methods("POST")
	router.HandleFunc("/admin/composer", s.adminOnly(s.handleComposerForm())).Methods("GET")
	router.HandleFunc("/admin/composer", s.adminOnly(s.handleComposerSubmit())).Methods("POST")
	router.HandleFunc("/admin/composer/media", s.adminOnly(s.handleMediaGallery())).Methods("GET")
	router.HandleFunc("/admin/composer/media", s.adminOnly(s.handleAddMediaToComposer())).Methods("POST")
	router.HandleFunc("/admin/composer/media/detail", s.adminOnly(s.handleMediaDetail())).Methods("GET")
	router.HandleFunc("/admin/composer/location", s.adminOnly(s.handleLocationSearch())).Methods("GET")
	router.HandleFunc("/admin/composer/location", s.adminOnly(s.handleAddLocationToComposer())).Methods("POST")
	router.HandleFunc("/admin/media/delete", s.adminOnly(s.handleDeleteMedia())).Methods("POST")
}

func (s Server) handleMicropubQuery() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// TODO
		accessToken, ok := r.Context().Value(contextKeyAccessToken).(app.TokenResponse)
		if !ok {
			s.logger.Error("failed fetch access token from context")
		}
		s.logger.WithField("access_token", accessToken).Info("found access token")

		switch r.URL.Query().Get("q") {
		case "source":
			limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
			if limit == 0 {
				limit = 30
			}
			after := r.URL.Query().Get("after")
			posts, err := s.App.QueryPostList(limit, after)
			if err != nil {
				s.logger.WithError(err).Error("failed to query post list")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			buf := bytes.NewBuffer([]byte{})
			err = json.NewEncoder(buf).Encode(posts.PostList)
			if err != nil {
				s.logger.WithError(err).Error("failed to encode post list")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Add("Content-type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(buf.Bytes())
		}
	}
}

func (s Server) handleMicropubCommand() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// TODO
		accessToken, ok := r.Context().Value(contextKeyAccessToken).(app.TokenResponse)
		if !ok {
			s.logger.Error("failed fetch access token from context")
		}

		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			s.logger.WithError(err).Error("failed to read request body")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.Body.Close()
		action := s.parser.ParseMicropubPostAction(bodyBytes)
		s.logger.
			WithField("access_token", accessToken).
			WithField("action", action).Info("micropub cmd!")

		switch action {
		case "create":
		case "update":
			updateRequest := s.parser.ParseUpdateRequest(bodyBytes)
			err := s.App.UpdatePost(updateRequest)
			if err != nil {
				s.logger.WithError(err).Error("failed to update post")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}

func (s Server) handleHomepage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// fetch latest posts
		limit := 12
		after := r.URL.Query().Get("after")
		postList, err := s.App.QueryPostList(limit, after)
		if err != nil {
			s.logger.WithError(err).Error("failed to query post list")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// view.render
		outBuf := new(bytes.Buffer)
		err = renderHomepage(outBuf, postList.PostList, postList.AfterKey)
		if err != nil {
			s.logger.WithError(err).Error("failed to render homepage")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-type", "text/html; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(outBuf.Bytes())
		if err != nil {
			s.logger.WithError(err).Error("failed to write html")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		return
	}
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

func (s Server) withBearerToken(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bearerToken := r.Header.Get("Authorization")
		if bearerToken == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		tokenEndpoint := os.Getenv("TOKEN_ENDPOINT")
		accessToken, err := s.App.VerifyAccessToken(tokenEndpoint, bearerToken)
		if err != nil {
			s.logger.WithError(err).Error("failed to verify access token")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if accessToken.IsValid() == false {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		ctx := context.WithValue(r.Context(), contextKeyAccessToken, accessToken)
		h(w, r.WithContext(ctx))
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

func (s Server) handleDeleteMedia() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		mediaURL := r.FormValue("media_url")
		err := s.App.DeleteMedia(mediaURL)
		if err != nil {
			s.logger.
				WithError(err).
				WithField("media_url", mediaURL).
				Error("failed to delete media")
			return
		}

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

		// TODO fetch session from context
		sess, ok := r.Context().Value(contextKeySessionID).(app.SessionData)
		if !ok {
			s.logger.Error("failed fetch session from context")
		}

		locationQuery := r.URL.Query().Get("location")
		locations := s.App.SearchLocations(sess.Location, locationQuery)
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
