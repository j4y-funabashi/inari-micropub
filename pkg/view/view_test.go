package view_test

import (
	"testing"
	"time"

	"github.com/j4y_funabashi/inari-micropub/pkg/app"
	"github.com/j4y_funabashi/inari-micropub/pkg/view"
	"github.com/matryer/is"
)

func TestParseMediaGallery(t *testing.T) {
	var tests = []struct {
		name string
	}{
		{name: "it works"},
	}

	for _, tt := range tests {
		is := is.NewRelaxed(t)
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			now := time.Now()

			// arrange
			presenter := view.NewPresenter()
			mediaRes := app.ShowMediaResponse{
				Media: []app.Media{
					app.Media{
						URL:      "test1",
						DateTime: &now,
					},
				},
				Years: []app.Year{
					app.Year{
						Year:           "2019",
						Count:          100,
						PublishedCount: 10,
					},
				},
				Months: []app.Month{
					app.Month{
						Month:          "1",
						Count:          10,
						PublishedCount: 1,
					},
				},
				Days: []app.Day{
					app.Day{
						Day:            "10",
						Count:          20,
						PublishedCount: 1,
					},
				},
				CurrentYear: app.Year{
					Year:           "2019",
					Count:          100,
					PublishedCount: 10,
				},
				CurrentMonth: app.Month{
					Month:          "1",
					Count:          10,
					PublishedCount: 1,
				},
				CurrentDay: app.Day{
					Day:            "10",
					Count:          20,
					PublishedCount: 1,
				},
			}

			expected := view.MediaGalleryView{
				Media: [][]view.Media{
					[]view.Media{
						view.Media{
							URL: "test1",
						},
					},
				},
				Years: []view.ProgressLink{
					view.ProgressLink{
						Name:  "2019",
						Value: 10,
						Total: 100,
						URL:   "?year=2019",
					},
				},
				Months: []view.ProgressLink{
					view.ProgressLink{
						Name:  "January",
						Value: 1,
						Total: 10,
						URL:   "?month=1&year=2019",
					},
				},
				Days: []view.ProgressLink{
					view.ProgressLink{
						Name:  "Thu 10",
						Value: 1,
						Total: 20,
						URL:   "?day=10&month=1&year=2019",
					},
				},
				CurrentDay: view.ProgressLink{
					Name:  "Thursday, 10 Jan 2019",
					Value: 1,
					Total: 20,
				},
			}

			// act
			result := presenter.ParseMediaGallery(mediaRes)

			// assert
			is.Equal(expected.Media, result.Media)
			is.Equal(expected.Years, result.Years)
			is.Equal(expected.Months, result.Months)
			is.Equal(expected.Days, result.Days)
			is.Equal(expected.CurrentDay, result.CurrentDay)
		})
	}
}
