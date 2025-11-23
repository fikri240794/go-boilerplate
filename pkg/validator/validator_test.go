package validator

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/fikri240794/gocerr"
	"github.com/stretchr/testify/assert"
)

type testValidStruct struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"required,min=1,max=150"`
}

type testStructWithJSONDash struct {
	IgnoreField string `json:"-" validate:"required"`
	ValidField  string `json:"valid_field" validate:"required"`
}

type testStructNoValidation struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type testStructWithOmitEmpty struct {
	Name     string `json:"name" validate:"required"`
	Nickname string `json:"nickname,omitempty" validate:"omitempty,min=3"`
}

func Test_fieldFromJSONTag(t *testing.T) {
	tests := []struct {
		name     string
		field    reflect.StructField
		expected string
	}{
		{
			name: "field with simple json tag should return tag value",
			field: reflect.StructField{
				Name: "TestField",
				Tag:  `json:"test_field"`,
			},
			expected: "test_field",
		},
		{
			name: "field with json tag and omitempty should return tag value",
			field: reflect.StructField{
				Name: "TestField",
				Tag:  `json:"test_field,omitempty"`,
			},
			expected: "test_field",
		},
		{
			name: "field with json dash should return empty string",
			field: reflect.StructField{
				Name: "IgnoredField",
				Tag:  `json:"-"`,
			},
			expected: "",
		},
		{
			name: "field with no json tag should return empty string",
			field: reflect.StructField{
				Name: "NoJSONTag",
				Tag:  ``,
			},
			expected: "",
		},
		{
			name: "field with json tag and multiple options should return tag value",
			field: reflect.StructField{
				Name: "TestField",
				Tag:  `json:"test_field,omitempty,string"`,
			},
			expected: "test_field",
		},
		{
			name: "field with empty json tag should return empty string",
			field: reflect.StructField{
				Name: "EmptyTag",
				Tag:  `json:""`,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fieldFromJSONTag(tt.field)

			assert.Equal(t, tt.expected, result, "Expected '%s'", tt.expected)
		})
	}
}

func Test_translateErrorToEnglish(t *testing.T) {
	t.Run("should initialize translator without panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			translateErrorToEnglish()
		}, "translateErrorToEnglish panicked")

		assert.NotNil(t, translator, "Expected translator to be initialized")
	})
}

func TestValidateStruct(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
		errorFields []string
	}{
		{
			name: "valid struct should return nil",
			input: testValidStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   30,
			},
			expectError: false,
		},
		{
			name: "missing required field should return error",
			input: testValidStruct{
				Name:  "",
				Email: "john@example.com",
				Age:   30,
			},
			expectError: true,
			errorFields: []string{"name"},
		},
		{
			name: "invalid email format should return error",
			input: testValidStruct{
				Name:  "John Doe",
				Email: "invalid-email",
				Age:   30,
			},
			expectError: true,
			errorFields: []string{"email"},
		},
		{
			name: "age below minimum should return error",
			input: testValidStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   0,
			},
			expectError: true,
			errorFields: []string{"age"},
		},
		{
			name: "age above maximum should return error",
			input: testValidStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   200,
			},
			expectError: true,
			errorFields: []string{"age"},
		},
		{
			name: "multiple validation errors should return all errors",
			input: testValidStruct{
				Name:  "",
				Email: "invalid-email",
				Age:   0,
			},
			expectError: true,
			errorFields: []string{"name", "email", "age"},
		},
		{
			name: "struct without validation tags should return nil",
			input: testStructNoValidation{
				Name:  "",
				Email: "",
			},
			expectError: false,
		},
		{
			name: "struct with omitempty and empty value should return nil",
			input: testStructWithOmitEmpty{
				Name:     "John",
				Nickname: "",
			},
			expectError: false,
		},
		{
			name: "struct with omitempty and invalid value should return error",
			input: testStructWithOmitEmpty{
				Name:     "John",
				Nickname: "ab",
			},
			expectError: true,
			errorFields: []string{"nickname"},
		},
		{
			name: "empty struct should return error for required fields",
			input: testValidStruct{
				Name:  "",
				Email: "",
				Age:   0,
			},
			expectError: true,
			errorFields: []string{"name", "email", "age"},
		},
		{
			name: "struct with json dash missing ValidField should return error",
			input: testStructWithJSONDash{
				IgnoreField: "some value",
				ValidField:  "",
			},
			expectError: true,
			errorFields: []string{"valid_field"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)

			if tt.expectError {
				assert.NotNil(t, err, "Expected error, got nil")

				customError, ok := gocerr.Parse(err)
				assert.True(t, ok, "Expected custom error, got %T", err)

				assert.Equal(t, http.StatusBadRequest, customError.Code, "Expected status code %d", http.StatusBadRequest)
				assert.Equal(t, http.StatusText(http.StatusBadRequest), customError.Message, "Expected message '%s'", http.StatusText(http.StatusBadRequest))

				if len(tt.errorFields) > 0 {
					assert.Len(t, customError.ErrorFields, len(tt.errorFields), "Expected %d error fields", len(tt.errorFields))

					fieldMap := make(map[string]bool)
					for _, ef := range customError.ErrorFields {
						fieldMap[ef.Field] = true
					}

					for _, expectedField := range tt.errorFields {
						assert.True(t, fieldMap[expectedField], "Expected error field '%s' not found", expectedField)
					}
				}
			} else {
				assert.Nil(t, err, "Expected no error, got %v", err)
			}
		})
	}
}
