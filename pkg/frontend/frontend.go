package frontend

import (
	"bytes"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct{}

func NewServer() Server {
	return Server{}
}

func (s Server) Routes(router *mux.Router) {
	router.HandleFunc("/", s.handleHomepage())
}

func (s Server) handleHomepage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// view.render
		outBuf := new(bytes.Buffer)
		err := RenderHomepage(outBuf)
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
