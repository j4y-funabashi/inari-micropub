package app_test

import (
	"testing"
	"time"

	"github.com/j4y_funabashi/inari-micropub/pkg/app"
	"github.com/j4y_funabashi/inari-micropub/pkg/eventlog"
	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
	"github.com/matryer/is"
	"github.com/sirupsen/logrus"
)

type mockEventlog struct{}

func (el mockEventlog) Append(ev eventlog.Event) error {
	return nil
}

func newEventlog() mockEventlog {
	return mockEventlog{}
}

type mockGeocoder struct {
}

func (g mockGeocoder) Lookup(address string) []app.Location {
	return []app.Location{}
}

func (g mockGeocoder) LookupLatLng(lat, lng float64) []app.Location {
	return []app.Location{}
}

func (g mockGeocoder) LookupVenues(lat, lng float64) []app.Location {
	return []app.Location{}
}

func (g mockGeocoder) SearchVenues(query string, lat, lng float64) []app.Location {
	return []app.Location{}
}

func newGeocoder() mockGeocoder {
	return mockGeocoder{}
}

type mockSelecta struct {
	years     []app.Year
	months    []app.Month
	days      []app.Day
	mediaList []app.Media
	media     app.Media
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

func (s mockSelecta) SelectMediaDayList(yr, month string) ([]app.Day, error) {
	return s.days, nil
}

func (s mockSelecta) SelectMediaMonthList(yr string) ([]app.Month, error) {
	return s.months, nil
}

func (s mockSelecta) SelectMediaYearList() []app.Year {
	return s.years
}

func (s mockSelecta) SelectMediaDay(yr, mn, d string) ([]app.Media, error) {
	return s.mediaList, nil
}

func (s mockSelecta) SelectMediaByURL(mediaURL string) (app.Media, error) {
	return s.media, nil
}

func (s mockSelecta) SelectPostList(limit int, afterKey string) mf2.PostList {
	return mf2.PostList{}
}

func (s mockSelecta) SelectPostByURL(uid string) (mf2.MicroFormat, error) {
	return mf2.MicroFormat{}, nil
}

func newMockSelecta(years []app.Year, months []app.Month) mockSelecta {
	return mockSelecta{
		years:  years,
		months: months,
	}
}

type mockSessionStore struct{}

func (ss mockSessionStore) Create() (app.SessionData, error) {
	return app.SessionData{}, nil
}

func (ss mockSessionStore) Fetch(sessionID string) (app.SessionData, error) {
	return app.SessionData{}, nil
}

func (ss mockSessionStore) Save(sess app.SessionData) error {
	return nil
}

func newMockSessionStore() mockSessionStore {
	return mockSessionStore{}
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
				CurrentYear: stubYear2,
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
				CurrentYear: stubYear1,
			},
		},
		{
			name:      "there are no years",
			y:         "",
			yearStubs: []app.Year{},
			expected: app.ShowMediaResponse{
				Years:       []app.Year{},
				CurrentYear: app.Year{},
			},
		},
	}

	for _, tt := range tests {

		is := is.NewRelaxed(t)
		tt := tt
		selecta := newMockSelecta(tt.yearStubs, []app.Month{})
		sessionStore := newMockSessionStore()
		geoCoder := newGeocoder()
		logger := logrus.New()
		el := newEventlog()

		t.Run(tt.name, func(t *testing.T) {
			// arrange
			sut := app.New(selecta, logger, sessionStore, geoCoder, el)

			// act
			result := sut.ShowMediaGallery(tt.y, tt.m, tt.d)

			// assert
			is.Equal(tt.expected.Years, result.Years)
			is.Equal(tt.expected.CurrentYear, result.CurrentYear)
		})
	}
}

func TestSessionToMf2(t *testing.T) {

	now := time.Now()

	var tests = []struct {
		name string
		sess app.SessionData
	}{
		{
			name: "empty session",
			sess: app.SessionData{},
		},
		{
			name: "full valid session",
			sess: app.SessionData{
				Media: []app.Media{
					app.Media{
						URL: "https://test1.jpg",
					},
					app.Media{
						URL: "https://test2.jpg",
					},
				},
				Location: app.Location{
					Lat:      6.66,
					Lng:      5.55,
					Locality: "meanwood",
					Region:   "West Yorkshire",
					Country:  "uk",
				},
				Published: &now,
				Content:   "hello test",
			},
		},
	}

	for _, tt := range tests {
		result := tt.sess.ToMf2()
		t.Errorf("%+v", result)
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
				CurrentMonth: stubMonth1,
			},
		},
	}

	logger := logrus.New()

	for _, tt := range tests {

		is := is.NewRelaxed(t)
		tt := tt
		selecta := newMockSelecta([]app.Year{}, tt.monthStubs)
		geoCoder := newGeocoder()
		sessionStore := newMockSessionStore()
		el := newEventlog()

		t.Run(tt.name, func(t *testing.T) {
			// arrange
			sut := app.New(selecta, logger, sessionStore, geoCoder, el)

			// act
			result := sut.ShowMediaGallery(tt.y, tt.m, tt.d)

			// assert
			is.Equal(tt.expected.Months, result.Months)
			is.Equal(tt.expected.CurrentMonth, result.CurrentMonth)
		})
	}
}
