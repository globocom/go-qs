package goqs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeQueryFormats(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected map[string]interface{}
	}{
		{
			name:     "Empty query",
			query:    "",
			expected: map[string]interface{}{},
		},
		{
			name:     "Simple string parameter",
			query:    "name=John",
			expected: map[string]interface{}{"name": "John"},
		},
		{
			name:     "Multiple parameters",
			query:    "a=1&b=hello&c=true",
			expected: map[string]interface{}{"a": "1", "b": "hello", "c": "true"},
		},
		{
			name:     "Array parameters",
			query:    "arr[]=1&arr[]=2",
			expected: map[string]interface{}{"arr": []interface{}{"1", "2"}},
		},
		{
			name:     "Nested object parameters",
			query:    "user[name]=Alice&user[age]=30",
			expected: map[string]interface{}{"user": map[string]interface{}{"name": "Alice", "age": "30"}},
		},
		{
			name:     "Numeric index parameters",
			query:    "a[1]=1&a[2]=2&a[3]=3",
			expected: map[string]interface{}{"a": []interface{}{"1", "2", "3"}},
		},
		{
			name:     "Numeric index parameters with empty indexes",
			query:    "a[]=1&a[]=2&a[]=3",
			expected: map[string]interface{}{"a": []interface{}{"1", "2", "3"}},
		},
		{
			name:     "Numeric index parameters with some empty index",
			query:    "a[0]=1&a[1]=2&a[]=3",
			expected: map[string]interface{}{"a": []interface{}{"1", "2", "3"}},
		},
		{
			name:     "Empty value parameter",
			query:    "empty=",
			expected: map[string]interface{}{"empty": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Parse(tt.query, &ParseOptions{
				AllowDots:                false,
				AllowEmptyArrays:         false,
				AllowPrototypes:          false,
				ParseNumbers:             false,
				AllowSparse:              false,
				ArrayLimit:               20,
				Charset:                  "utf-8",
				CharsetSentinel:          false,
				Comma:                    false,
				DecodeDotInKeys:          false,
				Decoder:                  nil,
				Delimiter:                "&",
				Depth:                    10,
				Duplicates:               "combine",
				IgnoreQueryPrefix:        false,
				InterpretNumericEntities: false,
				ParameterLimit:           1000,
				ParseArrays:              true,
				PlainObjects:             false,
				StrictDepth:              false,
				StrictNullHandling:       true,
				AllowNilArrayValues:      false,
				ThrowOnLimitExceeded:     false,
			})
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, res)
		})
	}
}
