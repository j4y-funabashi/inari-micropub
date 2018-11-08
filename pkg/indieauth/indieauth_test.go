package indieauth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/j4y_funabashi/inari-micropub/pkg/indieauth"
	"github.com/matryer/is"
	"github.com/sirupsen/logrus"
)

func TestVerifyToken(t *testing.T) {

	is := is.NewRelaxed(t)

	var tests = []struct {
		name        string
		token       string
		expectedErr error
		expectedRes indieauth.TokenResponse
	}{
		{
			name:        "happy path",
			expectedErr: nil,
			token:       "testtoken",
			expectedRes: indieauth.TokenResponse{
				Me:         "https://user.example.net/",
				ClientID:   "https://app.example.com/",
				Scope:      "create update delete",
				StatusCode: 200,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			// arrange
			tokenServer := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						is.Equal(
							"Bearer "+tt.token,
							r.Header.Get("Authorization"),
						)
						is.Equal(
							"application/json",
							r.Header.Get("Accept"),
						)
						jsonres := struct {
							Me       string `json:"me"`
							ClientID string `json:"client_id"`
							Scope    string `json:"scope"`
						}{
							Me:       "https://user.example.net/",
							ClientID: "https://app.example.com/",
							Scope:    "create update delete",
						}
						json.NewEncoder(w).Encode(jsonres)
						w.WriteHeader(200)
					},
				),
			)
			logger := logrus.New()

			// act
			result, err := indieauth.VerifyAccessToken(
				tokenServer.URL,
				tt.token,
				logger,
			)

			// assert
			is.Equal(tt.expectedRes, result)
			is.Equal(tt.expectedErr, err)

		})
	}
}
