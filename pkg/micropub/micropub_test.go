package micropub_test

import (
	"strings"
	"testing"

	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
	"github.com/j4y_funabashi/inari-micropub/pkg/micropub"
	"github.com/matryer/is"
	"github.com/sirupsen/logrus"
)

func TestMicropubQuery(t *testing.T) {
	var tests = []struct {
		name string
	}{
		{""},
	}

	is := is.NewRelaxed(t)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			logger := logrus.New()
			mediaURL := "https://j4y.co/micropub/media"
			mpServer := micropub.NewServer(
				mediaURL,
				logger,
				func(mf mf2.PostCreatedEvent) error { return nil },
			)

			// act
			result := mpServer.QueryConfig()

			// assert
			is.Equal(`{"media-endpoint":"https://j4y.co/micropub/media"}`, strings.TrimSpace(result.Body))
			is.Equal(200, result.StatusCode)
			is.Equal(result.Headers, map[string]string{"Content-type": "application/json"})
		})
	}
}

func TestCreatePost(t *testing.T) {
	var tests = []struct {
		name          string
		authorURL     string
		contentType   string
		body          string
		expectedEvent mf2.PostCreatedEvent
	}{
		{
			name:        "form values: happy path",
			authorURL:   "https://jay.funabashi.co.uk/",
			contentType: "application/x-www-form-urlencoded",
			body:        "uid=horse&published=2012-01-28",
			expectedEvent: mf2.PostCreatedEvent{
				EventData: mf2.MicroFormat{
					Type: []string{"h-entry"},
					Properties: map[string][]interface{}{
						"author":    []interface{}{"https://jay.funabashi.co.uk/"},
						"uid":       []interface{}{"horse"},
						"published": []interface{}{"2012-01-28"},
						"url":       []interface{}{"http://example.com/p/horse"},
					},
				},
			},
		},
		{
			name:        "json: happy path",
			authorURL:   "https://jay.funabashi.co.uk/",
			contentType: "application/json",
			body:        `{"type":["h-entry"], "properties": {"published": ["2012-01-28"], "uid": ["horse"]}}`,
			expectedEvent: mf2.PostCreatedEvent{
				EventData: mf2.MicroFormat{
					Type: []string{"h-entry"},
					Properties: map[string][]interface{}{
						"author":    []interface{}{"https://jay.funabashi.co.uk/"},
						"uid":       []interface{}{"horse"},
						"published": []interface{}{"2012-01-28"},
						"url":       []interface{}{"http://example.com/p/horse"},
					},
				},
			},
		},
	}

	is := is.NewRelaxed(t)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			logger := logrus.New()
			mediaURL := ""

			baseURL := "http://example.com/"
			authToken := "testtoken"
			createPostCount := 0
			createPostMock := func(event mf2.PostCreatedEvent) error {
				createPostCount++
				is.Equal(tt.expectedEvent.EventData, event.EventData)
				is.Equal("PostCreated", event.EventType)
				return nil
			}

			mpServer := micropub.NewServer(
				mediaURL,
				logger,
				createPostMock,
			)

			// act
			result := mpServer.CreatePost(
				baseURL,
				authToken,
				tt.contentType,
				tt.authorURL,
				tt.body,
			)

			// assert
			is.Equal(1, createPostCount)
			is.Equal(micropub.HttpResponse{StatusCode: 202, Headers: map[string]string{"Location": "http://example.com/p/horse"}}, result)
		})
	}
}
