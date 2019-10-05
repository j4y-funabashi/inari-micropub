package db_test

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	"github.com/j4y_funabashi/inari-micropub/pkg/db"
	"github.com/j4y_funabashi/inari-micropub/pkg/eventlog"
	"github.com/j4y_funabashi/inari-micropub/pkg/indieauth"
	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
	"github.com/j4y_funabashi/inari-micropub/pkg/micropub"
	"github.com/j4y_funabashi/inari-micropub/pkg/s3"
	"github.com/matryer/is"
	"github.com/sirupsen/logrus"
)

func TestSelectMediaList(t *testing.T) {
	t.SkipNow()

	var tests = []struct {
		name           string
		limit          int
		after          string
		expectedItems  []mf2.MediaMetadata
		expectedPaging mf2.ListPaging
	}{
		{
			name:  "it contains paging when there are more items",
			limit: 1,
			expectedItems: []mf2.MediaMetadata{
				mf2.MediaMetadata{URL: "http://example.com/2", IsPublished: true},
			},
			expectedPaging: mf2.ListPaging{
				After: "12345",
			},
		},
		{
			name:  "it does not contain paging when there are no more items",
			limit: 2,
			expectedItems: []mf2.MediaMetadata{
				mf2.MediaMetadata{URL: "http://example.com/2", IsPublished: true},
				mf2.MediaMetadata{URL: "http://example.com/1", IsPublished: false},
			},
			expectedPaging: mf2.ListPaging{},
		},
		{
			name:  "it pages items based on 'after' variable",
			limit: 1,
			expectedItems: []mf2.MediaMetadata{
				mf2.MediaMetadata{URL: "http://example.com/1", IsPublished: false},
			},
			expectedPaging: mf2.ListPaging{},
			after:          "12345",
		},
	}

	for _, tt := range tests {

		is := is.NewRelaxed(t)
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			// ARRANGE
			// setup DB
			sqlClient, err := db.OpenDB()
			if err != nil {
				t.Fatalf("failed to open DB: %s", err.Error())
			}

			// insert data
			_, err = sqlClient.Exec(
				`INSERT INTO media (id, year, month, sort_key, data) VALUES ($1, $2, $3, $4, $5)`,
				"http://example.com/1",
				"1984",
				"01",
				"1234",
				`{"url": "http://example.com/1"}`,
			)
			if err != nil {
				t.Fatalf("failed to insert data:: %s", err.Error())
			}
			// _, err = sqlClient.Exec(
			// 	`INSERT INTO media (id, year, month, sort_key, data) VALUES (:id, :year, :month, :sortkey, :data)`,
			// 	sql.Named("id", "http://example.com/2"),
			// 	sql.Named("year", "1984"),
			// 	sql.Named("month", "01"),
			// 	sql.Named("sortkey", "12345"),
			// 	sql.Named("data", `{"url": "http://example.com/2"}`),
			// )
			// _, err = sqlClient.Exec(
			// 	`INSERT INTO media_published (id) VALUES (:id)`,
			// 	sql.Named("id", "http://example.com/2"),
			// )
			// if err != nil {
			// 	t.Fatalf("failed to insert data:: %s", err.Error())
			// }

			// ACT
			selecta := db.NewSelecta(sqlClient)
			result, err := selecta.SelectMediaList(tt.limit, tt.after)
			if err != nil {
				t.Fatal("failed to select media list:: %s", err.Error())
			}

			// ASSERT
			is.Equal(tt.expectedItems, result.Items)
			is.Equal(tt.expectedPaging, *result.Paging)
		})

	}
}

func TestSelectPostList(t *testing.T) {
	t.SkipNow()

	var tests = []struct {
		name           string
		limit          int
		after          string
		expectedItems  []mf2.MicroFormat
		expectedPaging mf2.ListPaging
	}{
		{
			name:  "it contains paging when there are more items",
			limit: 1,
			expectedItems: []mf2.MicroFormat{
				mf2.MicroFormat{Properties: map[string][]interface{}{
					"uid": []interface{}{"http://example.com/2"},
				}},
			},
			expectedPaging: mf2.ListPaging{
				After: "12345",
			},
		},
		{
			name:  "it does not contain paging when there are no more items",
			limit: 2,
			expectedItems: []mf2.MicroFormat{
				mf2.MicroFormat{Properties: map[string][]interface{}{
					"uid": []interface{}{"http://example.com/2"},
				}},
				mf2.MicroFormat{Properties: map[string][]interface{}{
					"uid": []interface{}{"http://example.com/1"},
				}},
			},
			expectedPaging: mf2.ListPaging{},
		},
		{
			name:  "it pages items based on 'after' variable",
			limit: 1,
			expectedItems: []mf2.MicroFormat{
				mf2.MicroFormat{Properties: map[string][]interface{}{
					"uid": []interface{}{"http://example.com/1"},
				}},
			},
			expectedPaging: mf2.ListPaging{},
			after:          "12345",
		},
	}

	for _, tt := range tests {

		is := is.NewRelaxed(t)
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			// ARRANGE
			// setup DB
			sqlClient, err := db.OpenDB()
			if err != nil {
				t.Fatal("failed to open DB")
			}

			// insert data
			_, err = sqlClient.Exec(
				`INSERT INTO posts (id, year, month, sort_key, data) VALUES (:id, :year, :month, :sortkey, :data)`,
				sql.Named("id", "http://example.com/1"),
				sql.Named("year", "1984"),
				sql.Named("month", "01"),
				sql.Named("sortkey", "1234"),
				sql.Named("data", `{"properties": {"uid": ["http://example.com/1"]}}`),
			)
			if err != nil {
				t.Fatalf("failed to insert data:: %s", err.Error())
			}
			_, err = sqlClient.Exec(
				`INSERT INTO posts (id, year, month, sort_key, data) VALUES (:id, :year, :month, :sortkey, :data)`,
				sql.Named("id", "http://example.com/2"),
				sql.Named("year", "1984"),
				sql.Named("month", "01"),
				sql.Named("sortkey", "12345"),
				sql.Named("data", `{"properties": {"uid": ["http://example.com/2"]}}`),
			)
			if err != nil {
				t.Fatalf("failed to insert data:: %s", err.Error())
			}

			// ACT
			selecta := db.NewSelecta(sqlClient)
			result := selecta.SelectPostList(tt.limit, tt.after)
			if err != nil {
				t.Fatal("failed to select post list: ")
			}

			// ASSERT
			is.Equal(tt.expectedItems, result.Items)
			is.Equal(tt.expectedPaging, *result.Paging)
		})

	}
}

func buildTestMpServer(t *testing.T) micropub.Server {

	mediaEndpoint := ""
	tokenEndpoint := ""
	logger := logrus.New()
	verifyToken := func(tokenEndpoint, bearerToken string, logger *logrus.Logger) (indieauth.TokenResponse, error) {
		return indieauth.TokenResponse{}, nil
	}

	sqlClient, err := db.OpenDB()
	if err != nil {
		t.Fatal("failed to open DB")
	}
	selecta := db.NewSelecta(sqlClient)

	s3KeyPrefix := "test"
	s3Bucket := "test.events"
	s3Endpoint := "localstack:4572"
	s3Client, err := s3.NewClient(s3Endpoint)
	if err != nil {
		t.Fatal("failed to create s3 client")
	}

	err = s3Client.CreateBucket(s3Bucket)
	if err != nil {
		t.Errorf("failed to create s3 bucket: %s", err.Error())
	}

	eventLog := eventlog.NewEventLog(
		s3KeyPrefix,
		s3Bucket,
		s3Client,
		sqlClient,
		logger,
	)
	mediaServer := micropub.MediaServer{}

	mpserver := micropub.NewServer(
		mediaEndpoint,
		tokenEndpoint,
		logger,
		verifyToken,
		selecta,
		eventLog,
		mediaServer,
	)

	return mpserver
}

func TestCreateThenQuery(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.SkipNow()
	}

	// ARRANGE
	is := is.NewRelaxed(t)

	mpserver := buildTestMpServer(t)

	baseURL := ""
	contentType := "application/json"
	body := `{"properties": {"uid": ["test123"], "published": ["2019-01-28T13:13:13+00:00"], "photo": ["testphoto1.jpg"]}}`
	tokenRes := indieauth.TokenResponse{}

	// ACT
	createResponse, err := mpserver.CreatePost(
		baseURL,
		contentType,
		body,
		tokenRes,
	)
	if err != nil {
		t.Errorf("failed to create post: %s", err.Error())
	}

	byURLResult := mpserver.QuerySourceByURL(createResponse.Location)
	is.Equal(
		strings.TrimSpace(byURLResult.Body),
		`{"type":["h-entry"],"properties":{"author":[""],"photo":["testphoto1.jpg"],"published":["2019-01-28T13:13:13+00:00"],"uid":["test123"],"url":["/p/test123"]}}`,
	)
}

func TestSelectMediaList1(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.SkipNow()
	}

	// ARRANGE
	is := is.NewRelaxed(t)

	sqlClient, err := db.OpenDB()
	if err != nil {
		t.Fatal("failed to open DB")
	}
	selecta := db.NewSelecta(sqlClient)

	// ACT
	result := selecta.SelectMediaYearList()
	t.Logf("%+v", result)
	is.Equal("HORSE!", "BADGER!")
}
