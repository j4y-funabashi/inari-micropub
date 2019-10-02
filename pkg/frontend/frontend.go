package frontend

import (
	"bytes"
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
	router.HandleFunc("/", s.handleHomepage())
}

func (s Server) handleHomepage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// fetch latest posts
		limit := 12
		after := ""
		postList, err := s.App.QueryPostList(limit, after)
		if err != nil {
			s.logger.WithError(err).Error("failed to query post list")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// view.render
		outBuf := new(bytes.Buffer)
		err = RenderHomepage(outBuf, postList.PostList, postList.AfterKey)
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
