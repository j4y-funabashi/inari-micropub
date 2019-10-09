package app

import (
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

type Selecta interface {
	SelectMediaYearList() []Year
	SelectMediaMonthList(currentYear string) ([]Month, error)
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
	CurrentYear  string
	CurrentMonth string
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
	return ShowMediaResponse{
		Years:        years,
		CurrentYear:  currentYear,
		Months:       months,
		CurrentMonth: currentMonth,
	}
}

func parseCurrentMonth(selectedMonth string, months []Month) string {
	for _, yr := range months {
		if yr.Month == selectedMonth {
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
