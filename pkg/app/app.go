// Package app orchestrates application level actions
package app

import (
	"time"

	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
	"github.com/sirupsen/logrus"
)

type Server struct {
	selecta Selecta
	logger  *logrus.Logger
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
	SelectPostList(limit int, afterKey string) mf2.PostList
}

func New(selecta Selecta, logger *logrus.Logger) Server {
	return Server{
		selecta: selecta,
		logger:  logger,
	}
}

type ShowMediaResponse struct {
	Years        []Year
	Months       []Month
	Days         []Day
	CurrentYear  string
	CurrentMonth string
	CurrentDay   string
	Media        []Media
}

// ShowMedia fetches years and determines current year
func (s Server) ShowMedia(selectedYear, selectedMonth, selectedDay string) ShowMediaResponse {
	years := s.selecta.SelectMediaYearList()
	currentYear := parseCurrentYear(selectedYear, years)

	months, err := s.selecta.SelectMediaMonthList(currentYear)
	if err != nil {
		s.logger.WithError(err).Error("failed to select month list")
	}
	currentMonth := parseCurrentMonth(selectedMonth, months)

	days, err := s.selecta.SelectMediaDayList(currentYear, currentMonth)
	if err != nil {
		s.logger.WithError(err).Error("failed to select day list")
	}
	currentDay := parseCurrentDay(selectedDay, days)

	media, err := s.selecta.SelectMediaDay(currentYear, currentMonth, currentDay)
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

func parseCurrentDay(selectedDay string, days []Day) string {
	for _, day := range days {
		if day.Day == selectedDay {
			return selectedDay
		}
	}
	if len(days) > 0 {
		return days[0].Day
	}
	return ""
}

func parseCurrentMonth(selectedMonth string, months []Month) string {
	for _, month := range months {
		if month.Month == selectedMonth {
			return selectedMonth
		}
	}
	if len(months) > 0 {
		return months[0].Month
	}
	return ""
}

func parseCurrentYear(selectedYear string, years []Year) string {
	for _, yr := range years {
		if yr.Year == selectedYear {
			return selectedYear
		}
	}
	if len(years) > 0 {
		return years[0].Year
	}
	return ""
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
