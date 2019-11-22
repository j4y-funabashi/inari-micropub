package web_test

import (
	"testing"

	"github.com/j4y_funabashi/inari-micropub/pkg/app"
	"github.com/j4y_funabashi/inari-micropub/pkg/web"
	"github.com/matryer/is"
)

func TestParseMicropubAction(t *testing.T) {

	var tests = []struct {
		name        string
		requestBody string
		expected    string
	}{
		{
			name:        "empty values",
			requestBody: "",
			expected:    "",
		},
		{
			name:        "valid update json",
			requestBody: `{"action": "update", "url": "example.com"}`,
			expected:    "update",
		},
		{
			name:        "empty action",
			requestBody: `{"action": "", "url": "example.com"}`,
			expected:    "",
		},
	}

	for _, tt := range tests {
		is := is.NewRelaxed(t)
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			// act
			sut := web.NewParser()
			result := sut.ParseMicropubPostAction([]byte(tt.requestBody))

			// assert
			is.Equal(result, tt.expected)
		})
	}
}

func TestParseUpdateRequest(t *testing.T) {

	var tests = []struct {
		name        string
		requestBody string
		expected    app.UpdatePostRequest
	}{
		{
			name:        "empty values",
			requestBody: "",
			expected: app.UpdatePostRequest{
				Properties: map[string][]interface{}{},
			},
		},
		{
			name:        "valid replace",
			requestBody: `{"url": "https://example.com/post/100", "replace": {"content": ["hello moon"]}}`,
			expected: app.UpdatePostRequest{
				URL:  "https://example.com/post/100",
				Type: "replace",
				Properties: map[string][]interface{}{
					"content": []interface{}{
						"hello moon",
					},
				},
			},
		},
		{
			name: "valid replace - nested location",
			requestBody: `{
"url": "https://example.com/post/100",
"replace":
{
"location": [{"type": ["h-card"], "properties": {"city": ["leeds"]}}]
}

}`,
			expected: app.UpdatePostRequest{
				URL:  "https://example.com/post/100",
				Type: "replace",
				Properties: map[string][]interface{}{
					"location": []interface{}{
						map[string]interface{}{
							"type": []interface{}{"h-card"},
							"properties": map[string]interface{}{
								"city": []interface{}{
									"leeds",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		is := is.NewRelaxed(t)
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			// act
			sut := web.NewParser()
			result := sut.ParseUpdateRequest([]byte(tt.requestBody))

			// assert
			is.Equal(result, tt.expected)
		})
	}
}
