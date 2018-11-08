package micropub_test

import (
	"strings"
	"testing"

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
			logger := logrus.New()
			mediaURL := "https://j4y.co/micropub/media"
			mpServer := micropub.NewServer(mediaURL, logger)

			result := mpServer.QueryConfig()

			is.Equal(`{"media-endpoint":"https://j4y.co/micropub/media"}`, strings.TrimSpace(result.Body))
			is.Equal(200, result.StatusCode)
			is.Equal(result.Headers, map[string]string{"Content-type": "application/json"})
		})
	}
}
