package micropub_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestFormEncodedMicropub(t *testing.T) {

	var tests = []struct {
		name string
		data map[string]string
	}{
		{
			name: "create: happy path",
			data: map[string]string{
				"uid": "testuuid",
			},
		},
	}

	endpointURL := "http://mpserver/"
	contentType := "application/x-www-form-urlencoded"
	validToken := "test-valid-token"
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			// arrange
			is := is.NewRelaxed(t)

			httpclient := &http.Client{}
			data := url.Values{}
			for k, v := range tt.data {
				data.Set(k, v)
			}

			req, err := http.NewRequest(
				"POST",
				endpointURL,
				strings.NewReader(data.Encode()),
			)
			if err != nil {
				t.Fatalf("failed to build Request %s", err.Error())
			}
			req.Header.Add("Content-Type", contentType)
			req.Header.Add("Authorization", validToken)

			// act
			resp, err := httpclient.Do(req)
			if err != nil {
				t.Fatalf("failed to make Request %s", err.Error())
			}

			// assert
			is.Equal(202, resp.StatusCode)
			is.True(resp.Header.Get("Location") != "")

		})
	}
}
