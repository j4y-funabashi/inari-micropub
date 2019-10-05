package app_test

import (
	"testing"

	"github.com/j4y_funabashi/inari-micropub/pkg/app"
	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
	"github.com/matryer/is"
)

type mockSelecta struct {
	years []app.Year
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

func (s mockSelecta) SelectMediaYearList() []app.Year {
	return s.years
}

func (s mockSelecta) SelectPostList(limit int, afterKey string) mf2.PostList {
	return mf2.PostList{}
}

func newMockSelecta(years []app.Year) mockSelecta {
	return mockSelecta{
		years: years,
	}
}

func TestShowMedia(t *testing.T) {
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
		selecta := newMockSelecta(tt.yearStubs)

		t.Run(tt.name, func(t *testing.T) {
			sut := app.New(selecta)

			result := sut.ShowMedia(tt.y, tt.m, tt.d)

			is.Equal(tt.expected, result)
		})
	}
}
