package test_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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

func TestCreatePhoto(t *testing.T) {
	url := "http://localhost:3040/micropub/media"
	filePath := "photo-1.jpg"
	req := newFileUploadRequest(t, url, filePath, "test-valid-token")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		t.Log(err.Error())
		t.Fatal("failed to create photo")
	}
	t.Fatalf(
		"media endpoint returned: %d %s",
		res.StatusCode,
		res.Header.Get("Location"))
}

func newFileUploadRequest(t *testing.T, url, path, token string) *http.Request {
	file, err := os.Open(path)
	if err != nil {
		t.Log(err.Error())
		t.Fatal("failed to open file")
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		t.Log(err.Error())
		t.Fatal("failed to create form file")
	}
	_, err = io.Copy(part, file)
	if err != nil {
		t.Log(err.Error())
		t.Fatal("failed to copy file")
	}
	err = writer.Close()
	if err != nil {
		t.Log(err.Error())
		t.Fatal("failed to close writer")
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		t.Log(err.Error())
		t.Fatal("failed to create request")
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Add("Authorization", token)
	return req
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
