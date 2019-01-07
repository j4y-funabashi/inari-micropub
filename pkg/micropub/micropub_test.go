package micropub_test

import (
	"strings"
	"testing"

	"github.com/j4y_funabashi/inari-micropub/pkg/indieauth"
	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
	"github.com/j4y_funabashi/inari-micropub/pkg/micropub"
	"github.com/matryer/is"
	"github.com/sirupsen/logrus"
)

func TestConfigQuery(t *testing.T) {
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
			tokenEndpoint := "tokens.test.example.com"
			verifyTokenMock := func(tokenEndpoint, bearerToken string, logger *logrus.Logger) (indieauth.TokenResponse, error) {
				return indieauth.TokenResponse{Me: "https://jay.funabashi.co.uk/"}, nil
			}
			postList := mf2.PostList{}

			mpServer := micropub.NewServer(
				mediaURL,
				tokenEndpoint,
				logger,
				func(mf mf2.PostCreatedEvent) error { return nil },
				verifyTokenMock,
				&postList,
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
		tokenRes      indieauth.TokenResponse
		contentType   string
		body          string
		expectedEvent mf2.PostCreatedEvent
	}{
		{
			name:        "form values: happy path",
			tokenRes:    indieauth.TokenResponse{Me: "https://jay.funabashi.co.uk/"},
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
			tokenRes:    indieauth.TokenResponse{Me: "https://jay.funabashi.co.uk/"},
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
			tokenEndpoint := "tokens.test.example.com"
			createPostCount := 0
			createPostMock := func(event mf2.PostCreatedEvent) error {
				createPostCount++
				is.Equal(tt.expectedEvent.EventData, event.EventData)
				is.Equal("PostCreated", event.EventType)
				return nil
			}
			verifyTokenMock := func(tokenEndpoint, bearerToken string, logger *logrus.Logger) (indieauth.TokenResponse, error) {
				return indieauth.TokenResponse{Me: "https://jay.funabashi.co.uk/"}, nil
			}
			postList := mf2.PostList{}

			mpServer := micropub.NewServer(
				mediaURL,
				tokenEndpoint,
				logger,
				createPostMock,
				verifyTokenMock,
				&postList,
			)

			// act
			result := mpServer.CreatePost(
				baseURL,
				tt.contentType,
				tt.body,
				tt.tokenRes,
			)

			// assert
			is.Equal(1, createPostCount)
			is.Equal(micropub.HttpResponse{StatusCode: 202, Headers: map[string]string{"Location": "http://example.com/p/horse"}}, result)
		})
	}
}

func TestSourceQuery(t *testing.T) {

	var tests = []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "happy",
			url:      "https://example.com/1",
			expected: `{"type":["h-entry"],"properties":{"url":["https://example.com/1"]}}`,
		},
	}

	is := is.NewRelaxed(t)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			// arrange
			logger := logrus.New()
			mediaURL := ""

			tokenEndpoint := "tokens.test.example.com"
			createPostMock := func(event mf2.PostCreatedEvent) error { return nil }
			verifyTokenMock := func(tokenEndpoint, bearerToken string, logger *logrus.Logger) (indieauth.TokenResponse, error) {
				return indieauth.TokenResponse{Me: "https://jay.funabashi.co.uk/"}, nil
			}
			postList := mf2.PostList{}

			item1 := mf2.MicroFormat{
				Type: []string{"h-entry"},
				Properties: map[string][]interface{}{
					"url": []interface{}{tt.url},
				},
			}
			postList.Add(item1)

			mpServer := micropub.NewServer(
				mediaURL,
				tokenEndpoint,
				logger,
				createPostMock,
				verifyTokenMock,
				&postList,
			)

			result := mpServer.QuerySource(tt.url)
			is.Equal(tt.expected, strings.TrimSpace(result.Body))
		})
	}
}
