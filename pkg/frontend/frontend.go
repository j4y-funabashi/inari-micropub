package frontend

import (
	"fmt"
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
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", "jay.funabashi")
		return
	}
}
