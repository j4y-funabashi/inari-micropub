package frontend

import (
	"bytes"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/j4y_funabashi/inari-micropub/pkg/app"
)

type Server struct {
	App app.Server
}

func NewServer(a app.Server) Server {
	return Server{
		App: a,
	}
}

func (s Server) Routes(router *mux.Router) {
	router.HandleFunc("/", s.handleHomepage())
}

func (s Server) handleHomepage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// fetch latest posts
		limit := 10
		after := ""
		postList, err := s.App.QueryPostList(limit, after)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// view.render
		outBuf := new(bytes.Buffer)
		err = RenderHomepage(outBuf, postList.PostList)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-type", "text/html; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(outBuf.Bytes())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		return
	}
}
