package micropub_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestMicropub(t *testing.T) {
	var tests = []struct {
		name string
	}{
		{"HORSAEHIDREERCHICKEN"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			is := is.NewRelaxed(t)

			endpointURL := "http://mpserver/"

			httpclient := &http.Client{}
			data := url.Values{}
			data.Set("uid", "testuuid")
			req, err := http.NewRequest("POST", endpointURL, strings.NewReader(data.Encode()))
			if err != nil {
				t.Fatalf("failed to build Request %s", err.Error())
			}
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			resp, err := httpclient.Do(req)
			if err != nil {
				t.Fatalf("failed to make Request %s", err.Error())
			}

			is.Equal(202, resp.StatusCode)
			is.Equal("https://jay.funabashi.co.uk/p/testuuid", resp.Header.Get("Location"))

		})
	}
}
