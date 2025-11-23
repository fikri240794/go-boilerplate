package context

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name string
	Age  int
}

type contextKey string

const (
	key1ContextKey      contextKey = "key1"
	intKeyContextKey    contextKey = "intKey"
	boolKeyContextKey   contextKey = "boolKey"
	nonexistentKey      contextKey = "nonexistent"
	keyContextKey       contextKey = "key"
	structKeyContextKey contextKey = "structKey"
	ptrKeyContextKey    contextKey = "ptrKey"
	sliceKeyContextKey  contextKey = "sliceKey"
	mapKeyContextKey    contextKey = "mapKey"
)

func TestGetCtxValueSafely(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func() context.Context
		key            interface{}
		expectedResult interface{}
		validate       func(t *testing.T, result interface{})
	}{
		{
			name: "get string value successfully",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), key1ContextKey, "value1")
			},
			key:            key1ContextKey,
			expectedResult: "value1",
			validate: func(t *testing.T, result interface{}) {
				strResult, ok := result.(string)
				assert.True(t, ok, "Expected result to be string")
				assert.Equal(t, "value1", strResult, "Expected 'value1'")
			},
		},
		{
			name: "get int value successfully",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), intKeyContextKey, 12345)
			},
			key:            intKeyContextKey,
			expectedResult: 12345,
			validate: func(t *testing.T, result interface{}) {
				intResult, ok := result.(int)
				assert.True(t, ok, "Expected result to be int")
				assert.Equal(t, 12345, intResult, "Expected 12345")
			},
		},
		{
			name: "get bool value successfully",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), boolKeyContextKey, true)
			},
			key:            boolKeyContextKey,
			expectedResult: true,
			validate: func(t *testing.T, result interface{}) {
				boolResult, ok := result.(bool)
				assert.True(t, ok, "Expected result to be bool")
				assert.True(t, boolResult, "Expected true")
			},
		},
		{
			name: "return empty value when key not found",
			setupContext: func() context.Context {
				return context.Background()
			},
			key:            nonexistentKey,
			expectedResult: "",
			validate: func(t *testing.T, result interface{}) {
				strResult, ok := result.(string)
				assert.True(t, ok, "Expected result to be string")
				assert.Equal(t, "", strResult, "Expected empty string")
			},
		},
		{
			name: "return zero value for int when key not found",
			setupContext: func() context.Context {
				return context.Background()
			},
			key:            nonexistentKey,
			expectedResult: 0,
			validate: func(t *testing.T, result interface{}) {
				intResult, ok := result.(int)
				assert.True(t, ok, "Expected result to be int")
				assert.Equal(t, 0, intResult, "Expected 0")
			},
		},
		{
			name: "return empty value when type mismatch",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), keyContextKey, 123)
			},
			key:            keyContextKey,
			expectedResult: "",
			validate: func(t *testing.T, result interface{}) {
				strResult, ok := result.(string)
				assert.True(t, ok, "Expected result to be string")
				assert.Equal(t, "", strResult, "Expected empty string due to type mismatch")
			},
		},
		{
			name: "get struct value successfully",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), structKeyContextKey, TestStruct{Name: "John", Age: 30})
			},
			key:            structKeyContextKey,
			expectedResult: TestStruct{},
			validate: func(t *testing.T, result interface{}) {
				structResult, ok := result.(TestStruct)
				assert.True(t, ok, "Expected result to be TestStruct")
				assert.Equal(t, "John", structResult.Name, "Expected Name 'John'")
				assert.Equal(t, 30, structResult.Age, "Expected Age 30")
			},
		},
		{
			name: "get pointer value successfully",
			setupContext: func() context.Context {
				value := "pointer value"
				return context.WithValue(context.Background(), ptrKeyContextKey, &value)
			},
			key: ptrKeyContextKey,
			validate: func(t *testing.T, result interface{}) {
				ptrResult, ok := result.(*string)
				assert.True(t, ok, "Expected result to be *string")
				assert.NotNil(t, ptrResult, "Expected non-nil pointer")
				assert.Equal(t, "pointer value", *ptrResult, "Expected 'pointer value'")
			},
		},
		{
			name: "return nil pointer when key not found for pointer type",
			setupContext: func() context.Context {
				return context.Background()
			},
			key: nonexistentKey,
			validate: func(t *testing.T, result interface{}) {
				ptrResult, ok := result.(*string)
				assert.True(t, ok, "Expected result to be *string")
				assert.Nil(t, ptrResult, "Expected nil pointer")
			},
		},
		{
			name: "get slice value successfully",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), sliceKeyContextKey, []string{"a", "b", "c"})
			},
			key: sliceKeyContextKey,
			validate: func(t *testing.T, result interface{}) {
				sliceResult, ok := result.([]string)
				assert.True(t, ok, "Expected result to be []string")
				assert.Len(t, sliceResult, 3, "Expected slice length 3")
				assert.Equal(t, []string{"a", "b", "c"}, sliceResult, "Expected [a b c]")
			},
		},
		{
			name: "get map value successfully",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), mapKeyContextKey, map[string]int{"one": 1, "two": 2})
			},
			key: mapKeyContextKey,
			validate: func(t *testing.T, result interface{}) {
				mapResult, ok := result.(map[string]int)
				assert.True(t, ok, "Expected result to be map[string]int")
				assert.Len(t, mapResult, 2, "Expected map length 2")
				assert.Equal(t, map[string]int{"one": 1, "two": 2}, mapResult, "Expected map with values 1 and 2")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()

			switch tt.expectedResult.(type) {
			case string:
				result := GetCtxValueSafely[string](ctx, tt.key)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			case int:
				result := GetCtxValueSafely[int](ctx, tt.key)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			case bool:
				result := GetCtxValueSafely[bool](ctx, tt.key)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			case *string:
				result := GetCtxValueSafely[*string](ctx, tt.key)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			case []string:
				result := GetCtxValueSafely[[]string](ctx, tt.key)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			case map[string]int:
				result := GetCtxValueSafely[map[string]int](ctx, tt.key)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			case TestStruct:
				result := GetCtxValueSafely[TestStruct](ctx, tt.key)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}
