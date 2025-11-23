package grpc_metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestMDToMapString(t *testing.T) {
	tests := []struct {
		name     string
		input    metadata.MD
		expected map[string]string
	}{
		{
			name:     "empty metadata should return empty map",
			input:    metadata.MD{},
			expected: map[string]string{},
		},
		{
			name: "metadata with single key-value should return correct map",
			input: metadata.MD{
				"authorization": []string{"Bearer token123"},
			},
			expected: map[string]string{
				"authorization": "Bearer token123",
			},
		},
		{
			name: "metadata with multiple keys should return correct map",
			input: metadata.MD{
				"authorization": []string{"Bearer token123"},
				"user-id":       []string{"12345"},
				"request-id":    []string{"abc-def-ghi"},
			},
			expected: map[string]string{
				"authorization": "Bearer token123",
				"user-id":       "12345",
				"request-id":    "abc-def-ghi",
			},
		},
		{
			name: "metadata with multiple values should use first value only",
			input: metadata.MD{
				"authorization": []string{"Bearer token123", "Bearer token456", "Bearer token789"},
			},
			expected: map[string]string{
				"authorization": "Bearer token123",
			},
		},
		{
			name: "metadata with empty value slice should skip that key",
			input: metadata.MD{
				"authorization": []string{"Bearer token123"},
				"empty-key":     []string{},
				"user-id":       []string{"12345"},
			},
			expected: map[string]string{
				"authorization": "Bearer token123",
				"user-id":       "12345",
			},
		},
		{
			name: "metadata with mixed values should handle correctly",
			input: metadata.MD{
				"key1": []string{"value1"},
				"key2": []string{},
				"key3": []string{"value3a", "value3b"},
				"key4": []string{"value4"},
			},
			expected: map[string]string{
				"key1": "value1",
				"key3": "value3a",
				"key4": "value4",
			},
		},
		{
			name: "metadata with special characters should preserve them",
			input: metadata.MD{
				"content-type": []string{"application/json"},
				"x-api-key":    []string{"key@#$%^&*()"},
			},
			expected: map[string]string{
				"content-type": "application/json",
				"x-api-key":    "key@#$%^&*()",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MDToMapString(tt.input)

			assert.Len(t, result, len(tt.expected), "Expected map length %d", len(tt.expected))
			assert.Equal(t, tt.expected, result, "Result map should match expected map")
		})
	}
}

func TestMDGetString(t *testing.T) {
	tests := []struct {
		name     string
		input    metadata.MD
		key      string
		expected string
	}{
		{
			name:     "empty metadata should return empty string",
			input:    metadata.MD{},
			key:      "authorization",
			expected: "",
		},
		{
			name: "existing key should return first value",
			input: metadata.MD{
				"authorization": []string{"Bearer token123"},
			},
			key:      "authorization",
			expected: "Bearer token123",
		},
		{
			name: "non-existing key should return empty string",
			input: metadata.MD{
				"authorization": []string{"Bearer token123"},
			},
			key:      "user-id",
			expected: "",
		},
		{
			name: "key with multiple values should return first value only",
			input: metadata.MD{
				"authorization": []string{"Bearer token123", "Bearer token456", "Bearer token789"},
			},
			key:      "authorization",
			expected: "Bearer token123",
		},
		{
			name: "key with empty value slice should return empty string",
			input: metadata.MD{
				"authorization": []string{},
			},
			key:      "authorization",
			expected: "",
		},
		{
			name: "case insensitive key lookup should match (gRPC metadata keys are case-insensitive)",
			input: metadata.MD{
				"authorization": []string{"Bearer token123"},
			},
			key:      "Authorization",
			expected: "Bearer token123",
		},
		{
			name: "key with special characters should work",
			input: metadata.MD{
				"x-api-key": []string{"key@#$%^&*()"},
			},
			key:      "x-api-key",
			expected: "key@#$%^&*()",
		},
		{
			name: "key with empty string value should return empty string",
			input: metadata.MD{
				"empty-value": []string{""},
			},
			key:      "empty-value",
			expected: "",
		},
		{
			name: "key with whitespace value should preserve it",
			input: metadata.MD{
				"space-value": []string{"   value with spaces   "},
			},
			key:      "space-value",
			expected: "   value with spaces   ",
		},
		{
			name: "multiple keys with correct key should return correct value",
			input: metadata.MD{
				"authorization": []string{"Bearer token123"},
				"user-id":       []string{"12345"},
				"request-id":    []string{"abc-def-ghi"},
			},
			key:      "user-id",
			expected: "12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MDGetString(tt.input, tt.key)

			assert.Equal(t, tt.expected, result, "Expected '%s'", tt.expected)
		})
	}
}
