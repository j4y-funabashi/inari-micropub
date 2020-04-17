package app_test

import (
	"os"
	"testing"
	"time"

	"github.com/go-test/deep"
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

func TestSessionToMf2View(t *testing.T) {

	now, err := time.Parse("2006-01-02T15:04:05Z", "1984-01-28T10:10:10Z")
	if err != nil {
		t.Fatalf("failed to create now:: %s", err.Error())
	}

	var tests = []struct {
		name     string
		sess     app.SessionData
		expected mf2.MicroFormatView
	}{
		{
			name: "empty session",
			sess: app.SessionData{
				UID:       "test-uuid-123",
				Published: &now,
			},
			expected: mf2.MicroFormatView{
				Type:      "entry",
				Published: "1984-01-28T10:10:10Z",
				Uid:       "test-uuid-123",
				Url:       "/p/test-uuid-123",
				Archive:   "198401",
			},
		},
		{
			name: "full valid session",
			sess: app.SessionData{
				UID: "test-uuid-1234",
				Media: []app.Media{
					app.Media{
						URL: "https://test1.jpg",
					},
					app.Media{
						URL: "https://test2.jpg",
					},
				},
				Location: app.Location{
					Name:     "Nandos",
					Lat:      6.66,
					Lng:      5.55,
					Locality: "meanwood",
					Region:   "West Yorkshire",
					Country:  "uk",
				},
				Published: &now,
				Content:   "hello test",
			},
			expected: mf2.MicroFormatView{
				Type:      "entry",
				Published: "1984-01-28T10:10:10Z",
				Uid:       "test-uuid-1234",
				Url:       "/p/test-uuid-1234",
				Content:   "hello test",
				Photo:     []string{"https://test1.jpg", "https://test2.jpg"},
				Location:  "Nandos, meanwood, West Yorkshire, uk",
				Archive:   "198401",
			},
		},
	}

	for _, tc := range tests {
		is := is.NewRelaxed(t)
		result := tc.sess.ToMf2()
		t.Logf("TN ::: %s", tc.name)
		t.Logf("RS ::: %#v", result.ToView())
		t.Logf("EX ::: %#v", tc.expected)
		is.Equal(result.ToView(), tc.expected)
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

func TestExtractMediaMetadata(t *testing.T) {

	t1, err := time.Parse(time.RFC3339, "2019-11-09T01:57:37Z")
	if err != nil {
		t.Fatalf("failed to parse time: %s", err.Error())
	}

	var tests = []struct {
		name     string
		fileName string
		expected app.MediaMetadataResponse
	}{
		{
			name:     "works",
			fileName: "test_data/photo-1.jpg",
			expected: app.MediaMetadataResponse{
				FileHash: "6c5e11de40eda50688d2980613bb9181",
				MimeType: "image/jpeg",
				URL:      "test-domain/media/2019/6c5e11de40eda50688d2980613bb9181.jpg",
				FileKey:  "2019/6c5e11de40eda50688d2980613bb9181.jpg",
				DateTime: &t1,
				Lat:      53.79932783333333,
				Lng:      -1.538786861111111,
			},
		},
	}

	baseURL := "test-domain"
	for _, tt := range tests {
		// arrange
		file, err := os.Open(tt.fileName)
		if err != nil {
			t.Fatalf("failed to open file: %s", tt.fileName)
		}

		// act
		result, _ := app.ExtractMediaMetadata(file, tt.fileName, baseURL)

		// assert
		if diff := deep.Equal(tt.expected, result); diff != nil {
			t.Error(diff)
		}
	}
}
