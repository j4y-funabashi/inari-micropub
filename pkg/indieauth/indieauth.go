package indieauth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

func VerifyAccessToken(
	tokenEndpoint,
	bearerToken string,
	logger *logrus.Logger,
) (TokenResponse, error) {

	// build request
	req, err := http.NewRequest("GET", tokenEndpoint, nil)
	if err != nil {
		logger.Errorf("failed to create GET request: %s", err.Error())
		return TokenResponse{}, err
	}
	bearerToken = "Bearer " + strings.Replace(bearerToken, "Bearer", "", -1)
	req.Header.Add("Authorization", bearerToken)
	req.Header.Add("Accept", "application/json")

	// make request
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		logger.Errorf("failed to GET token endpoint: %s", err.Error())
		return TokenResponse{}, err
	}

	// read response
	tokenRes := TokenResponse{StatusCode: resp.StatusCode}
	buf := bytes.Buffer{}
	buf.ReadFrom(resp.Body)
	err = json.Unmarshal(buf.Bytes(), &tokenRes)
	if err != nil {
		logger.Errorf("failed to unmarshal response body: %v  %s", buf.String(), err.Error())
		return TokenResponse{}, err
	}
	return tokenRes, nil
}

type TokenResponse struct {
	Me               string `json:"me"`
	ClientID         string `json:"client_id"`
	Scope            string `json:"scope"`
	IssuedBy         string `json:"issued_by"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	StatusCode       int
}

func (tr TokenResponse) IsValid() bool {
	if tr.StatusCode != 200 {
		return false
	}
	if strings.TrimSpace(tr.Me) == "" {
		return false
	}
	if strings.TrimSpace(tr.Scope) == "" {
		return false
	}
	return true
}
