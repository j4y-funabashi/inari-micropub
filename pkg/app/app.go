// Package app orchestrates application level actions
package app

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/j4y_funabashi/inari-micropub/pkg/eventlog"
	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

type Server struct {
	selecta      Selecta
	logger       *logrus.Logger
	sessionStore SessionStore
	geo          Geocoder
	el           EventLog
}

type Geocoder interface {
	Lookup(address string) []Location
	LookupLatLng(lat, lng float64) []Location
	LookupVenues(lat, lng float64) []Location
	SearchVenues(query string, lat, lng float64) []Location
}

type Year struct {
	Year           string
	Count          int
	PublishedCount int
}

type Month struct {
	Month          string
	Count          int
	PublishedCount int
}

type Day struct {
	Day            string
	Count          int
	PublishedCount int
}

type Media struct {
	URL         string     `json:"url"`
	MimeType    string     `json:"mime_type"`
	DateTime    *time.Time `json:"date_time"`
	Lat         float64    `json:"lat"`
	Lng         float64    `json:"lng"`
	IsPublished bool       `json:"is_published"`
}

type Location struct {
	Name     string  `json:"name"`
	Lat      float64 `json:"latitude"`
	Lng      float64 `json:"longitude"`
	Locality string  `json:"locality"`
	Region   string  `json:"region"`
	Country  string  `json:"country-name"`
}

func (l Location) isValid() bool {
	if l.Lat == 0 && l.Lng == 0 {
		return false
	}
	return true
}

func (l Location) toGeoURL() string {
	return fmt.Sprintf("geo:%g,%g", l.Lat, l.Lng)
}

func (l Location) ToMf2() mf2.MicroFormat {
	mfType := []string{"h-adr"}
	if l.Name != "" {
		mfType = []string{"h-card"}
	}

	props := make(map[string][]interface{})
	if l.Name != "" {
		props["name"] = append(props["name"], l.Name)
	}
	props["latitude"] = append(props["latitude"], l.Lat)
	props["longitude"] = append(props["longitude"], l.Lng)
	props["locality"] = append(props["locality"], l.Locality)
	props["region"] = append(props["region"], l.Region)
	props["country-name"] = append(props["country-name"], l.Country)

	mf := mf2.MicroFormat{
		Type:       mfType,
		Properties: props,
	}

	return mf
}

type Selecta interface {
	SelectMediaYearList() []Year
	SelectMediaMonthList(currentYear string) ([]Month, error)
	SelectMediaDayList(currentYear, currentMonth string) ([]Day, error)
	SelectMediaDay(year, month, day string) ([]Media, error)
	SelectMediaByURL(url string) (Media, error)
	SelectPostList(limit int, afterKey string) mf2.PostList
}

type SessionStore interface {
	Create() (SessionData, error)
	Fetch(sessionID string) (SessionData, error)
	Save(sess SessionData) error
}

type EventLog interface {
	Append(event eventlog.Event) error
}

func New(
	selecta Selecta,
	logger *logrus.Logger,
	ss SessionStore,
	geo Geocoder,
	el EventLog,
) Server {
	return Server{
		selecta:      selecta,
		logger:       logger,
		sessionStore: ss,
		geo:          geo,
		el:           el,
	}
}

type ShowMediaResponse struct {
	Years        []Year
	Months       []Month
	Days         []Day
	CurrentYear  Year
	CurrentMonth Month
	CurrentDay   Day
	Media        []Media
}

type ShowMediaDetailResponse struct {
	Media Media
}

type SessionData struct {
	Token              string     `json:"token"`
	Media              []Media    `json:"media"`
	Location           Location   `json:"location"`
	SuggestedLocations []Location `json:"suggested_locations"`
	Published          *time.Time `json:"published"`
	Content            string     `json:"content"`
}

func (s SessionData) Reset() SessionData {
	return SessionData{
		Token: s.Token,
	}
}

func (s *SessionData) AddSuggestedLocations(locations []Location) {
	for _, loc := range locations {
		s.SuggestedLocations = append(s.SuggestedLocations, loc)
	}
}

func (s SessionData) CookieValue(maxAge int) string {
	return fmt.Sprintf("session_id=%s; Path=/; Max-Age=%d", s.Token, maxAge)
}

func (s SessionData) ToMf2() mf2.MicroFormat {
	uid := uuid.NewV4()
	mfType := []string{"h-entry"}
	now := time.Now()

	props := make(map[string][]interface{})
	props["uid"] = append(props["uid"], uid.String())

	// media
	for _, m := range s.Media {
		props["photo"] = append(props["photo"], m.URL)
	}

	// location
	if s.Location.isValid() {
		props["location"] = append(props["location"], s.Location.ToMf2())
	}

	// published
	if s.Published == nil {
		s.Published = &now
	}
	props["published"] = append(props["published"], s.Published.Format(time.RFC3339))

	// content
	if s.Content != "" {
		props["content"] = append(props["content"], s.Content)
	}

	baseURL := os.Getenv("SITE_URL")
	postURL := strings.TrimRight(baseURL, "/") + "/p/" + uid.String()

	mf := mf2.MicroFormat{
		Type:       mfType,
		Properties: props,
	}
	mf.SetDefaults(baseURL, uid.String(), postURL)

	return mf
}

type AuthResponse struct {
	Session SessionData
}

func (s Server) DeleteMedia(mediaURL string) error {
	event := eventlog.NewMediaDeleted(mediaURL)
	return s.el.Append(event)
}

func (s Server) CreatePost(sess SessionData) error {
	mf := sess.ToMf2()
	event := eventlog.NewPostCreated(mf)
	return s.el.Append(event)
}

func (s Server) FetchSession(sessionID string) (SessionData, error) {
	sess, err := s.sessionStore.Fetch(sessionID)
	return sess, err
}

func (s Server) SaveSession(sess SessionData) error {
	return s.sessionStore.Save(sess)
}

func (s Server) Auth(password string) (AuthResponse, error) {
	res := AuthResponse{}

	// compare hashes
	h := sha1.New()
	h.Write([]byte(password))
	actualPass := os.Getenv("ADMIN_PASSWORD")
	hashedPass := hex.EncodeToString(h.Sum(nil))
	if actualPass != hashedPass {
		return res, errors.New("incorrect password")
	}

	// create session
	session, err := s.sessionStore.Create()
	if err != nil {
		s.logger.WithError(err).Error("failed to create session")
		return res, err
	}
	res.Session = session

	return res, nil
}

func (s Server) SearchLocations(location Location, query string) []Location {
	if query == "" {
		return []Location{}
	}

	if location.isValid() {
		return s.geo.SearchVenues(query, location.Lat, location.Lng)
	}
	return s.geo.Lookup(query)
}

func (s Server) SearchLocationsByLatLng(lat, lng float64) []Location {
	if lat == 0 && lng == 0 {
		return []Location{}
	}
	locations := s.geo.LookupLatLng(lat, lng)
	return locations
}

func (s Server) SearchVenues(lat, lng float64) []Location {
	if lat == 0 && lng == 0 {
		return []Location{}
	}
	locations := s.geo.LookupVenues(lat, lng)
	return locations
}

func (s Server) ShowMediaDetail(mediaURL string) ShowMediaDetailResponse {
	out := ShowMediaDetailResponse{}

	media, err := s.selecta.SelectMediaByURL(mediaURL)
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch media")
		return out
	}

	out.Media = media

	return out
}

// ShowMediaGallery fetches years and determines current year
func (s Server) ShowMediaGallery(selectedYear, selectedMonth, selectedDay string) ShowMediaResponse {
	years := s.selecta.SelectMediaYearList()
	currentYear := parseCurrentYear(selectedYear, years)

	months, err := s.selecta.SelectMediaMonthList(currentYear.Year)
	if err != nil {
		s.logger.WithError(err).Error("failed to select month list")
	}
	currentMonth := parseCurrentMonth(selectedMonth, months)

	days, err := s.selecta.SelectMediaDayList(currentYear.Year, currentMonth.Month)
	if err != nil {
		s.logger.WithError(err).Error("failed to select day list")
	}
	currentDay := parseCurrentDay(selectedDay, days)

	media, err := s.selecta.SelectMediaDay(currentYear.Year, currentMonth.Month, currentDay.Day)
	if err != nil {
		s.logger.WithError(err).Error("failed to select media day")
	}

	return ShowMediaResponse{
		Years:        years,
		CurrentYear:  currentYear,
		Months:       months,
		CurrentMonth: currentMonth,
		Days:         days,
		CurrentDay:   currentDay,
		Media:        media,
	}
}

func parseCurrentDay(selectedDay string, days []Day) Day {
	for _, day := range days {
		if day.Day == selectedDay {
			return day
		}
	}
	if len(days) > 0 {
		return days[0]
	}
	return Day{}
}

func parseCurrentMonth(selectedMonth string, months []Month) Month {
	for _, month := range months {
		if month.Month == selectedMonth {
			return month
		}
	}
	if len(months) > 0 {
		return months[0]
	}
	return Month{}
}

func parseCurrentYear(selectedYear string, years []Year) Year {
	for _, yr := range years {
		if yr.Year == selectedYear {
			return yr
		}
	}
	if len(years) > 0 {
		return years[0]
	}
	return Year{}
}

type QueryPostListResponse struct {
	PostList []mf2.MicroFormatView
	AfterKey string
}

func (s Server) QueryPostList(limit int, after string) (*QueryPostListResponse, error) {
	pl := s.selecta.SelectPostList(limit, after)
	postList := []mf2.MicroFormatView{}

	for _, mf2 := range pl.Items {
		postList = append(postList, mf2.ToView())
	}

	return &QueryPostListResponse{
		PostList: postList,
		AfterKey: pl.Paging.After,
	}, nil
}
