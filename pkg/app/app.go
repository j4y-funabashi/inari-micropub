package app

import "github.com/j4y_funabashi/inari-micropub/pkg/mf2"

type Server struct {
	selecta Selecta
}

type Year struct {
	Year           string
	Count          int
	PublishedCount int
}

type Selecta interface {
	SelectMediaYearList() []Year
	SelectPostList(limit int, afterKey string) mf2.PostList
}

func New(selecta Selecta) Server {
	return Server{
		selecta: selecta,
	}
}

type ShowMediaResponse struct {
	Years       []Year
	CurrentYear string
}

// ShowMedia fetches years and determines current year
func (s Server) ShowMedia(selectedYear, selectedMonth, selectedDay string) ShowMediaResponse {
	years := s.selecta.SelectMediaYearList()
	currentYear := parseCurrentYear(selectedYear, years)
	return ShowMediaResponse{
		Years:       years,
		CurrentYear: currentYear,
	}
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
