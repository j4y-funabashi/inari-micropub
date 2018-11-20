package micropub

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

type Server struct {
	mediaEndpoint   string
	createPostEvent func(mf mf2.MicroFormat) error
	logger          *logrus.Logger
}

type HttpResponse struct {
	Body       string
	StatusCode int
	Headers    map[string]string
}

func NewServer(mediaEndpoint string, logger *logrus.Logger, createPost func(mf mf2.MicroFormat) error) Server {
	return Server{
		mediaEndpoint:   mediaEndpoint,
		createPostEvent: createPost,
		logger:          logger,
	}
}

func (s Server) Routes(router *mux.Router) {
	router.HandleFunc("/", s.handleMicropub())
}

func (s Server) handleMicropub() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		baseUrl := "https://jay.funabashi.co.uk/"
		authToken := r.Header.Get("Authorization")
		contentType := r.Header.Get("Content-Type")
		authorURL := "http:/jay.example.com"
		buf := bytes.Buffer{}
		buf.ReadFrom(r.Body)

		response := HttpResponse{}

		if r.Method == "GET" {
			switch r.URL.Query().Get("q") {
			case "config":
				response = s.QueryConfig()
			}
		}

		if r.Method == "POST" {

			response = s.CreatePost(
				baseUrl,
				authToken,
				contentType,
				authorURL,
				buf.String(),
			)
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

func (s Server) CreatePost(
	baseUrl,
	authToken,
	contentType,
	authorURL,
	body string,
) HttpResponse {

	// create uid and post permalink
	uid := uuid.NewV4()

	s.logger.
		WithField("content_type", contentType).
		Info("checking Content-type header")
	mf, err := s.buildMF(body, contentType)
	if err != nil {
		s.logger.WithError(err).Error("failed to build mf")
		return HttpResponse{
			StatusCode: http.StatusInternalServerError,
		}
	}

	uuid := uid.String()
	if mf.GetFirstString("uid") != "" {
		uuid = mf.GetFirstString("uid")
	}
	postUrl := strings.TrimRight(baseUrl, "/") + "/p/" + uuid
	mf.SetDefaults(authorURL, uuid, postUrl)
	s.logger.
		WithField("mf", mf).
		Info("mf built")

	err = s.createPostEvent(mf)
	if err != nil {
		s.logger.WithError(err).Error("failed to save post")
		return HttpResponse{
			StatusCode: http.StatusInternalServerError,
		}
	}

	headers := map[string]string{
		"Location": postUrl,
	}
	return HttpResponse{
		StatusCode: http.StatusAccepted,
		Headers:    headers,
	}
}

func (s Server) buildMF(body, contentType string) (mf2.MicroFormat, error) {
	var mf mf2.MicroFormat

	if contentType == "application/x-www-form-urlencoded" {
		parsedBody, err := url.ParseQuery(body)
		if err != nil {
			return mf, err
		}

		mf = mf2.MfFromForm(parsedBody)
		return mf, nil
	}

	if contentType == "application/json" {
		mf, err := mf2.MfFromJson(body)
		if err != nil {
			return mf, err
		}
		return mf, nil
	}

	return mf, nil
}
