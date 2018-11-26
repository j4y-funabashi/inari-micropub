package micropub

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/j4y_funabashi/inari-micropub/pkg/indieauth"
	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

type Server struct {
	mediaEndpoint   string
	tokenEndpoint   string
	createPostEvent func(mf mf2.PostCreatedEvent) error
	logger          *logrus.Logger
	verifyToken     func(tokenEndpoint, bearerToken string, logger *logrus.Logger) (indieauth.TokenResponse, error)
}

type HttpResponse struct {
	Body       string
	StatusCode int
	Headers    map[string]string
}

func NewServer(
	mediaEndpoint,
	tokenEndpoint string,
	logger *logrus.Logger,
	createPost func(event mf2.PostCreatedEvent) error,
	verifyToken func(tokenEndpoint, bearerToken string, logger *logrus.Logger) (indieauth.TokenResponse, error),
) Server {
	return Server{
		mediaEndpoint:   mediaEndpoint,
		tokenEndpoint:   tokenEndpoint,
		createPostEvent: createPost,
		logger:          logger,
		verifyToken:     verifyToken,
	}
}

func (s Server) Routes(router *mux.Router) {
	baseURL := "https://jay.funabashi.co.uk/"
	router.HandleFunc("/", s.handleMicropub(baseURL))
	router.HandleFunc("/health", s.handleHealthcheck())
}

func (s Server) handleHealthcheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func (s Server) handleMicropub(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		authToken := r.Header.Get("Authorization")
		contentType := r.Header.Get("Content-Type")

		tokenRes, err := s.verifyToken(
			s.tokenEndpoint,
			authToken,
			s.logger,
		)
		if err != nil {
			s.logger.WithError(err).Error("failed to verify token")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if tokenRes.IsValid() == false {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		response := HttpResponse{}

		if r.Method == "GET" {
			switch r.URL.Query().Get("q") {
			case "config":
				response = s.QueryConfig()
			}
		}

		if r.Method == "POST" {

			body := bytes.Buffer{}
			body.ReadFrom(r.Body)
			response = s.CreatePost(
				baseURL,
				contentType,
				body.String(),
				tokenRes,
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
	baseURL,
	contentType,
	body string,
	tokenRes indieauth.TokenResponse,
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
	postUrl := strings.TrimRight(baseURL, "/") + "/p/" + uuid
	mf.SetDefaults(tokenRes.Me, uuid, postUrl)
	s.logger.
		WithField("mf", mf).
		Info("mf built")

	event := mf2.NewPostCreated(mf)

	err = s.createPostEvent(event)
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
