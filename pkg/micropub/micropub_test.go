package micropub_test

import (
	"testing"

	"github.com/j4y_funabashi/inari-micropub/pkg/micropub"
	"github.com/matryer/is"
)

func TestItWorks(t *testing.T) {
	var tests = []struct {
		name     string
		body     string
		expected string
	}{
		{
			name:     "valid update",
			body:     `{"action": "update", "url": "blah.com"}`,
			expected: "update",
		},
		{
			name:     "empty json",
			body:     `{}`,
			expected: "create",
		},
		{
			name:     "create json",
			body:     `{"type": "h-entry", "content": "hellchicken"}`,
			expected: "create",
		},
		{
			name:     "create json",
			body:     `{"type": "h-entry", "content": "hellchicken"}`,
			expected: "create",
		},
		{
			name:     "form encoded",
			body:     `h=entry&content=hello+world`,
			expected: "create",
		},
	}

	for _, tt := range tests {
		is := is.NewRelaxed(t)
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			result := micropub.ParsePostAction(tt.body)
			is.Equal(result, tt.expected)
		})
	}
}
