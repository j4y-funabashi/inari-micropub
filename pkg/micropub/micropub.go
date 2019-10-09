package micropub

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/j4y_funabashi/inari-micropub/pkg/db"
	"github.com/j4y_funabashi/inari-micropub/pkg/eventlog"
	"github.com/j4y_funabashi/inari-micropub/pkg/indieauth"
	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
	"github.com/j4y_funabashi/inari-micropub/pkg/s3"
	"github.com/rwcarlsen/goexif/exif"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

// MediaServer contains methods for uploading and fetching files
type MediaServer struct {
	mediaEndpoint string
	s3Client      s3.Client
	s3Bucket      string
}

func NewMediaServer(s3Client s3.Client, endpoint string, bucket string) MediaServer {
	return MediaServer{
		mediaEndpoint: endpoint,
		s3Client:      s3Client,
		s3Bucket:      bucket,
	}
}

func (s MediaServer) UploadMedia(fileKey string, mediaFile io.Reader) error {
	err := s.s3Client.WriteObject(fileKey, s.s3Bucket, mediaFile, false)
	return err
}

func (s MediaServer) DownloadMedia(fileKey string) (*bytes.Buffer, error) {
	buf, err := s.s3Client.ReadObject(fileKey, s.s3Bucket)
	return buf, err
}

type Server struct {
	mediaEndpoint string
	tokenEndpoint string
	logger        *logrus.Logger
	verifyToken   func(tokenEndpoint, bearerToken string, logger *logrus.Logger) (indieauth.TokenResponse, error)
	selecta       db.Selecta
	eventLog      eventlog.EventLog
	mediaServer   MediaServer
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
	verifyToken func(tokenEndpoint, bearerToken string, logger *logrus.Logger) (indieauth.TokenResponse, error),
	selecta db.Selecta,
	eventLog eventlog.EventLog,
	mediaServer MediaServer,
) Server {
	return Server{
		mediaEndpoint: mediaEndpoint,
		tokenEndpoint: tokenEndpoint,
		logger:        logger,
		verifyToken:   verifyToken,
		selecta:       selecta,
		eventLog:      eventLog,
		mediaServer:   mediaServer,
	}
}

func (s Server) Routes(router *mux.Router) {
	baseURL := os.Getenv("BASE_URL")
	siteURL := os.Getenv("SITE_URL")

	router.HandleFunc("/health", s.handleHealthcheck())
	router.HandleFunc("/", s.handleMicropub(siteURL))
	router.HandleFunc("/media", s.handleMedia(baseURL)).Methods("POST")
	router.HandleFunc("/media", s.handleMediaQuery()).Methods("GET")
	router.HandleFunc("/media/{year}/{fileKey}", s.handleMediaDownload()).Methods("GET")
}

type UploadedFile struct {
	Filename string
	File     io.Reader
}

func (s Server) handleMediaQuery() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		response := HttpResponse{}

		switch r.URL.Query().Get("q") {
		case "source":
			url := r.URL.Query().Get("url")
			if len(url) > 0 {
				response = s.QueryMediaByURL(url)
			} else {
				limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
				if limit == 0 {
					limit = 20
				}
				after := r.URL.Query().Get("after")
				year := r.URL.Query().Get("year")
				month := r.URL.Query().Get("month")
				response = s.QueryMediaList(limit, after, year, month)
			}
		case "years":
			response = s.QueryMediaYearsList()
		case "months":
			year := r.URL.Query().Get("year")
			response = s.QueryMediaMonthsList(year)
		}

		for k, v := range response.Headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(response.StatusCode)
		w.Write([]byte(response.Body))

	}
}

func (s Server) handleMediaDownload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		s.logger.Info(vars)
		fileKey := path.Join(vars["year"], vars["fileKey"])
		s.logger.Info(fileKey)

		buf, err := s.mediaServer.DownloadMedia(fileKey)
		if err != nil {
			s.logger.WithError(err).Error("failed to download media file")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.ServeContent(w, r, fileKey, time.Now(), bytes.NewReader(buf.Bytes()))
	}
}

func (s Server) handleMedia(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			s.logger.WithError(err).Error("Failed to parse multipart form")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for _, mediaFile := range r.MultipartForm.File["file"] {

			// INIT METADATA
			fileExt := strings.ToLower(path.Ext(mediaFile.Filename))
			now := time.Now()
			uid := uuid.NewV4()
			mediaMeta := mf2.MediaMetadata{
				Uid:      uid.String(),
				DateTime: &now,
			}

			// OPEN FILE
			file, err := mediaFile.Open()
			if err != nil {
				s.logger.WithError(err).Error("failed to open file")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer file.Close()

			// HASH
			hash := md5.New()
			if _, err := io.Copy(hash, file); err != nil {
				s.logger.WithError(err).Error("Failed to hash file")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			hashInBytes := hash.Sum(nil)[:16]
			mediaMeta.FileHash = hex.EncodeToString(hashInBytes)

			// REWIND FILE
			_, err = file.Seek(0, 0)
			if err != nil {
				s.logger.WithError(err).Error("Failed to rewind file")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// GET MIME TYPE
			mimeType, err := fetchMimeType(file)
			if err != nil {
				s.logger.WithError(err).Error("Failed to fetch mime type")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			mediaMeta.MimeType = mimeType

			// REWIND FILE
			_, err = file.Seek(0, 0)
			if err != nil {
				s.logger.WithError(err).Error("Failed to rewind file")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if mimeType == "image/jpeg" {
				// READ EXIF
				exifData, err := fetchEXIF(file)
				if err != nil {
					s.logger.WithError(err).Error("Failed to fetch EXIF data")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				mediaMeta.DateTime = exifData.dateTime
				mediaMeta.Lat = exifData.lat
				mediaMeta.Lng = exifData.lng
			}

			// TODO make this work better, use url pkg
			fileKey := path.Join(
				mediaMeta.DateTime.Format("2006"),
				mediaMeta.FileHash,
			) + fileExt
			mediaMeta.FileKey = fileKey

			// TODO make this work better, use url pkg
			mediaURL := strings.TrimRight(baseURL, "/") + "/media/" + fileKey
			mediaMeta.URL = mediaURL

			s.logger.Info(mediaMeta)

			// REWIND FILE
			_, err = file.Seek(0, 0)
			if err != nil {
				s.logger.WithError(err).Error("Failed to rewind file")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// upload media file to s3
			err = s.mediaServer.UploadMedia(mediaMeta.FileKey, file)
			if err != nil {
				s.logger.WithError(err).Error("Failed to upload media to s3")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// append to eventlog
			event := eventlog.NewMediaUploaded(mediaMeta)
			err = s.eventLog.Append(event)
			if err != nil {
				s.logger.WithError(err).Error("Failed to append event to log")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("Location", mediaMeta.URL)
			w.WriteHeader(http.StatusCreated)
		}
	}
}

func fetchMimeType(file io.Reader) (string, error) {
	// get mime type
	buf := make([]byte, 512)
	_, err := file.Read(buf)
	if err != nil {
		return "", err
	}
	contentType := http.DetectContentType(buf)

	return contentType, nil
}

type exifData struct {
	dateTime *time.Time
	lat      float64
	lng      float64
}

func fetchEXIF(file io.Reader) (exifData, error) {

	out := exifData{}

	// GET EXIF DATA
	exifData, err := exif.Decode(file)
	if err != nil {
		return out, err
	}

	dateTime, err := exifData.DateTime()
	if err == nil {
		out.dateTime = &dateTime
	}

	lat, lng, err := exifData.LatLong()
	if err == nil {
		out.lat = lat
		out.lng = lng
	}

	return out, nil
}

func (s Server) handleHealthcheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", "healthy")
		return
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
			s.logger.WithField("tokenRes", tokenRes).Info("Invalid token")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		response := HttpResponse{}

		if r.Method == "GET" {
			switch r.URL.Query().Get("q") {
			case "config":
				response = s.QueryConfig()
			case "source":
				url := r.URL.Query().Get("url")
				if len(url) > 0 {
					response = s.QuerySourceByURL(url)
				} else {
					limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
					if limit == 0 {
						limit = 30
					}
					after := r.URL.Query().Get("after")
					response = s.QuerySourceList(limit, after)
				}
			case "years":
				response = s.QueryYearsList()
			case "months":
				year := r.URL.Query().Get("year")
				response = s.QueryMonthsList(year)
			}
		}

		if r.Method == "POST" {
			body := bytes.Buffer{}
			_, err := body.ReadFrom(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			action := ParsePostAction(body.String())

			switch action {
			case "create":
				createResponse, err := s.CreatePost(
					baseURL,
					contentType,
					body.String(),
					tokenRes,
				)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Header().Set("Location", createResponse.Location)
				w.WriteHeader(http.StatusAccepted)
				return
			case "update":
				s.UpdatePost(body.String())
				return
			}
		}

		for k, v := range response.Headers {
			w.Header().Set(k, v)
		}

		w.WriteHeader(response.StatusCode)
		w.Write([]byte(response.Body))
	}
}

func (s Server) UpdatePost(body string) {
	s.logger.
		WithField("body", body).
		Info("I CAN HAZ UPDATEXOIS")
}

func ParsePostAction(body string) string {

	e := struct {
		Action string `json:"action"`
	}{}
	err := json.Unmarshal([]byte(body), &e)
	if err != nil {
		return "create"
	}

	if e.Action != "" {
		return e.Action
	}
	return "create"
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

type CreatePostResponse struct {
	Location string
}

func (s Server) CreatePost(
	baseURL,
	contentType,
	body string,
	tokenRes indieauth.TokenResponse,
) (*CreatePostResponse, error) {

	// create post
	uid := uuid.NewV4()
	s.logger.
		WithField("content_type", contentType).
		Info("checking Content-type header")
	mf, err := s.buildMF(body, contentType)
	if err != nil {
		s.logger.WithError(err).Error("failed to build mf")
		return nil, err
	}
	uuid := uid.String()
	if mf.GetFirstString("uid") != "" {
		uuid = mf.GetFirstString("uid")
	}
	postURL := strings.TrimRight(baseURL, "/") + "/p/" + uuid
	mf.SetDefaults(tokenRes.Me, uuid, postURL)
	s.logger.
		WithField("mf", mf).
		Info("mf built")
	event := eventlog.NewPostCreated(mf)

	// add event to eventlog
	err = s.eventLog.Append(event)
	if err != nil {
		s.logger.WithError(err).Error("failed to save post")
		return nil, err
	}

	return &CreatePostResponse{
		Location: postURL,
	}, nil
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

func (s Server) QueryMediaByURL(url string) HttpResponse {

	body, err := s.selecta.SelectMediaByURL(url)
	if err != nil {
		return HttpResponse{
			Body: err.Error(),
		}
	}

	buf := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(buf).Encode(body)
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

func (s Server) QueryMediaList(limit int, after, year, month string) HttpResponse {

	var err error
	var body mf2.MediaList
	if month != "" && year != "" {
		body, err = s.selecta.SelectMediaMonth(year, month)
		if err != nil {
			return HttpResponse{
				Body: err.Error(),
			}
		}

	} else {
		body, err = s.selecta.SelectMediaList(limit, after)
		if err != nil {
			return HttpResponse{
				Body: err.Error(),
			}
		}

	}

	buf := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(buf).Encode(body)
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

func (s Server) QueryMonthsList(year string) HttpResponse {
	body, err := s.selecta.SelectMonthList(year)
	if err != nil {
		return HttpResponse{
			Body: err.Error(),
		}
	}

	buf := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(buf).Encode(body)
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

func (s Server) QueryYearsList() HttpResponse {
	body, err := s.selecta.SelectYearList()
	if err != nil {
		return HttpResponse{
			Body: err.Error(),
		}
	}

	buf := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(buf).Encode(body)
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

func (s Server) QueryMediaYearsList() HttpResponse {
	body := s.selecta.SelectMediaYearList()
	buf := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(buf).Encode(body)
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

func (s Server) QueryMediaMonthsList(year string) HttpResponse {
	body, err := s.selecta.SelectMediaMonthList(year)
	if err != nil {
		s.logger.WithError(err).Error("failed to select month list")
		return HttpResponse{
			Body: err.Error(),
		}
	}

	buf := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(buf).Encode(body)
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

func (s Server) QuerySourceList(limit int, after string) HttpResponse {
	body := s.selecta.SelectPostList(limit, after)
	buf := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(buf).Encode(body)
	if err != nil {
		return HttpResponse{
			Body:       err.Error(),
			StatusCode: http.StatusInternalServerError,
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

func (s Server) QuerySourceByURL(url string) HttpResponse {

	body, err := s.selecta.SelectPostByURL(url)
	if err != nil {
		return HttpResponse{
			Body: err.Error(),
		}
	}

	buf := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(buf).Encode(body)
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
