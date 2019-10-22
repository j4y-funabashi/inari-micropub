// Package app orchestrates application level actions
package app

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
	"github.com/sirupsen/logrus"
)

type Server struct {
	selecta      Selecta
	logger       *logrus.Logger
	sessionStore SessionStore
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
}

func New(selecta Selecta, logger *logrus.Logger, ss SessionStore) Server {
	return Server{
		selecta:      selecta,
		logger:       logger,
		sessionStore: ss,
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
	Token string `json:"token"`
}

func (s SessionData) CookieValue() string {
	return fmt.Sprintf("session_id=%s; Path=/", s.Token)
}

type AuthResponse struct {
	Session SessionData
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
