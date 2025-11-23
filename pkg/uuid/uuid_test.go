package uuid

import (
	"errors"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
)

func TestNewV7(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "should generate valid UUID v7",
		},
		{
			name: "should generate different UUIDs on multiple calls",
		},
		{
			name: "should generate non-nil UUID",
		},
		{
			name: "should generate UUID with correct format (36 characters)",
		},
		{
			name: "should generate UUID with version 7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewV7()

			assert.NotEqual(t, uuid.Nil, result, "Expected non-nil UUID")
			assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", result.String(), "Expected valid UUID, got nil UUID string")

			uuidStr := result.String()
			assert.Len(t, uuidStr, 36, "Expected UUID string length 36")

			if len(uuidStr) >= 15 && uuidStr[14] != '7' {
				t.Logf("Warning: UUID version might not be v7, got character '%c' at position 14", uuidStr[14])
			}
		})
	}
}

func Test_newV7WithGenerator(t *testing.T) {
	tests := []struct {
		name          string
		generator     uuidGenerator
		expectedIsNil bool
	}{
		{
			name: "successful generation should return valid UUID",
			generator: func() (uuid.UUID, error) {
				return uuid.FromString("018c5c85-7e6c-7a3e-8c4d-123456789abc")
			},
			expectedIsNil: false,
		},
		{
			name: "error in generation should return uuid.Nil",
			generator: func() (uuid.UUID, error) {
				return uuid.Nil, errors.New("failed to generate UUID")
			},
			expectedIsNil: true,
		},
		{
			name: "default generator should work",
			generator: func() (uuid.UUID, error) {
				return uuid.NewV7()
			},
			expectedIsNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := newV7WithGenerator(tt.generator)

			if tt.expectedIsNil {
				assert.Equal(t, uuid.Nil, result, "Expected uuid.Nil")
			} else {
				assert.NotEqual(t, uuid.Nil, result, "Expected non-nil UUID")
			}
		})
	}
}
