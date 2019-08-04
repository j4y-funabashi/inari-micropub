package db_test

import (
	"database/sql"
	"testing"

	"github.com/j4y_funabashi/inari-micropub/pkg/db"
	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
	"github.com/matryer/is"
)

func TestSelectMediaList(t *testing.T) {

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
				t.Fatal("failed to open DB")
			}

			// insert data
			_, err = sqlClient.Exec(
				`INSERT INTO media (id, year, month, sort_key, data) VALUES (:id, :year, :month, :sortkey, :data)`,
				sql.Named("id", "http://example.com/1"),
				sql.Named("year", "1984"),
				sql.Named("month", "01"),
				sql.Named("sortkey", "1234"),
				sql.Named("data", `{"url": "http://example.com/1"}`),
			)
			if err != nil {
				t.Fatalf("failed to insert data:: %s", err.Error())
			}
			_, err = sqlClient.Exec(
				`INSERT INTO media (id, year, month, sort_key, data) VALUES (:id, :year, :month, :sortkey, :data)`,
				sql.Named("id", "http://example.com/2"),
				sql.Named("year", "1984"),
				sql.Named("month", "01"),
				sql.Named("sortkey", "12345"),
				sql.Named("data", `{"url": "http://example.com/2"}`),
			)
			_, err = sqlClient.Exec(
				`INSERT INTO media_published (id) VALUES (:id)`,
				sql.Named("id", "http://example.com/2"),
			)
			if err != nil {
				t.Fatalf("failed to insert data:: %s", err.Error())
			}

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
			result, err := selecta.SelectPostList(tt.limit, tt.after)
			if err != nil {
				t.Fatal("failed to select post list: ")
			}

			// ASSERT
			is.Equal(tt.expectedItems, result.Items)
			is.Equal(tt.expectedPaging, *result.Paging)
		})

	}
}
