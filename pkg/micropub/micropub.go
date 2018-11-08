package micropub

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Server struct {
	mediaEndpoint string
}

type HttpResponse struct {
	Body       string
	StatusCode int
	Headers    map[string]string
}

func NewServer(mediaEndpoint string, logger *logrus.Logger) Server {
	return Server{
		mediaEndpoint: mediaEndpoint,
	}
}

func (s Server) Routes(router *mux.Router) {
	router.HandleFunc("/", s.HandleMicropub())
}

func (s Server) HandleMicropub() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := HttpResponse{}

		if r.Method == "GET" {
			switch r.URL.Query().Get("q") {
			case "config":
				response = s.QueryConfig()
			}
		}

		for k, v := range response.Headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(response.StatusCode)
		w.Write([]byte(response.Body))
	}
}

func (s Server) QueryConfig() HttpResponse {
	buf := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(buf).Encode(
		struct {
			MediaEndpoint string `json:"media-endpoint"`
		}{
			MediaEndpoint: s.mediaEndpoint},
	)
	if err != nil {
		return HttpResponse{
			Body: err.Error(),
		}
	}
	headers := map[string]string{
		"Content-type": "application/json",
	}
	return HttpResponse{
		Headers:    headers,
		Body:       buf.String(),
		StatusCode: http.StatusOK,
	}
}

type MpServer struct{}

func CreatePost() {
}
