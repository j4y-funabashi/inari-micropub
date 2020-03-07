package test_test

import (
	"bytes"
	"net/http"
	"testing"
)

func TestCreatePostWithNoToken(t *testing.T) {
	url := "http://localhost:3040/micropub"
	createPost(url, "", 401, t)
}

func TestCreatePost(t *testing.T) {
	url := "http://localhost:3040/micropub"
	createPost(url, "test-valid-token", 201, t)
}

func createPost(url, token string, expectedStatus int, t *testing.T) {
	buf := []byte(`{}`)
	requestBody := bytes.NewBuffer(buf)
	req, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		t.Log(err.Error())
		t.Fatal("failed to create request")
	}
	req.Header.Add("Authorization", token)

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		t.Log(err.Error())
		t.Fatal("failed to create post")
	}
	if res.StatusCode != expectedStatus {
		t.Errorf(
			"expected status code %d, got %d",
			expectedStatus,
			res.StatusCode)
	}
}
