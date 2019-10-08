package app_test

import (
	"testing"

	"github.com/j4y_funabashi/inari-micropub/pkg/app"
	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
	"github.com/matryer/is"
)

type mockSelecta struct {
	years  []app.Year
	months []app.Month
}

var stubYear1 = app.Year{
	Year:           "1984",
	Count:          1,
	PublishedCount: 10,
}
var stubYear2 = app.Year{
	Year:           "2001",
	Count:          10,
	PublishedCount: 100,
}

var stubMonth1 = app.Month{
	Month:          "1",
	Count:          1,
	PublishedCount: 10,
}
var stubMonth2 = app.Month{
	Month:          "12",
	Count:          10,
	PublishedCount: 100,
}

func (s mockSelecta) SelectMediaMonthList(yr string) []app.Month {
	return s.months
}

func (s mockSelecta) SelectMediaYearList() []app.Year {
	return s.years
}

func (s mockSelecta) SelectPostList(limit int, afterKey string) mf2.PostList {
	return mf2.PostList{}
}

func newMockSelecta(years []app.Year, months []app.Month) mockSelecta {
	return mockSelecta{
		years:  years,
		months: months,
	}
}

func TestShowMediaYears(t *testing.T) {
	var tests = []struct {
		name      string
		y         string
		m         string
		d         string
		yearStubs []app.Year
		expected  app.ShowMediaResponse
	}{
		{
			name: "current year is selected year",
			y:    "2001",
			yearStubs: []app.Year{
				stubYear1,
				stubYear2,
			},
			expected: app.ShowMediaResponse{
				Years: []app.Year{
					stubYear1,
					stubYear2,
				},
				CurrentYear: "2001",
			},
		},
		{
			name: "current year is first with empty selected year",
			y:    "",
			yearStubs: []app.Year{
				stubYear1,
				stubYear2,
			},
			expected: app.ShowMediaResponse{
				Years: []app.Year{
					stubYear1,
					stubYear2,
				},
				CurrentYear: "1984",
			},
		},
		{
			name:      "there are no years",
			y:         "",
			yearStubs: []app.Year{},
			expected: app.ShowMediaResponse{
				Years:       []app.Year{},
				CurrentYear: "",
			},
		},
	}

	for _, tt := range tests {

		is := is.NewRelaxed(t)
		tt := tt
		selecta := newMockSelecta(tt.yearStubs, []app.Month{})

		t.Run(tt.name, func(t *testing.T) {
			// arrange
			sut := app.New(selecta)

			// act
			result := sut.ShowMedia(tt.y, tt.m, tt.d)

			// assert
			is.Equal(tt.expected.Years, result.Years)
			is.Equal(tt.expected.CurrentYear, result.CurrentYear)
		})
	}
}

func TestShowMediaMonths(t *testing.T) {
	var tests = []struct {
		name       string
		y          string
		m          string
		d          string
		monthStubs []app.Month
		expected   app.ShowMediaResponse
	}{
		{
			name: "current month is selected month",
			m:    "1",
			monthStubs: []app.Month{
				stubMonth1,
				stubMonth2,
			},
			expected: app.ShowMediaResponse{
				Months: []app.Month{
					stubMonth1,
					stubMonth2,
				},
				CurrentMonth: "1",
			},
		},
	}

	for _, tt := range tests {

		is := is.NewRelaxed(t)
		tt := tt
		selecta := newMockSelecta([]app.Year{}, tt.monthStubs)

		t.Run(tt.name, func(t *testing.T) {
			// arrange
			sut := app.New(selecta)

			// act
			result := sut.ShowMedia(tt.y, tt.m, tt.d)

			// assert
			is.Equal(tt.expected.Months, result.Months)
			is.Equal(tt.expected.CurrentMonth, result.CurrentMonth)
		})
	}
}
