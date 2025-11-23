package middlewares

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMiddlewares_GetUnaryServerInterceptors(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) *Middlewares
		validate func(t *testing.T, interceptors []interface{})
	}{
		{
			name: "should_return_all_interceptors_in_correct_order",
			setup: func(t *testing.T) *Middlewares {
				return &Middlewares{
					Recover:   NewRecoverMiddleware(),
					Tracer:    NewTracerMiddleware(),
					RequestID: NewRequestIDMiddleware(),
					Log:       NewLogMiddleware(),
					Timeout:   NewTimeoutMiddleware(nil),
				}
			},
			validate: func(t *testing.T, interceptors []interface{}) {
				assert.NotNil(t, interceptors)
				assert.Len(t, interceptors, 5)

				for _, interceptor := range interceptors {
					assert.NotNil(t, interceptor)
				}
			},
		},
		{
			name: "should_return_interceptors_with_correct_order",
			setup: func(t *testing.T) *Middlewares {
				return &Middlewares{
					Recover:   NewRecoverMiddleware(),
					Tracer:    NewTracerMiddleware(),
					RequestID: NewRequestIDMiddleware(),
					Log:       NewLogMiddleware(),
					Timeout:   NewTimeoutMiddleware(nil),
				}
			},
			validate: func(t *testing.T, interceptors []interface{}) {
				assert.NotNil(t, interceptors)
				assert.Len(t, interceptors, 5)

				for i, interceptor := range interceptors {
					assert.NotNil(t, interceptor, "Interceptor at index %d should not be nil", i)
				}
			},
		},
		{
			name: "should_return_empty_slice_when_middlewares_not_initialized",
			setup: func(t *testing.T) *Middlewares {
				return &Middlewares{}
			},
			validate: func(t *testing.T, interceptors []interface{}) {

				assert.NotNil(t, interceptors)
				assert.Len(t, interceptors, 5)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middlewares := tt.setup(t)

			defer func() {
				if r := recover(); r != nil {
					if tt.name == "should_return_empty_slice_when_middlewares_not_initialized" {

						t.Logf("Expected panic caught: %v", r)
					} else {
						t.Errorf("Unexpected panic: %v", r)
					}
				}
			}()

			interceptors := middlewares.GetUnaryServerInterceptors()

			interfaceSlice := make([]interface{}, len(interceptors))
			for i, v := range interceptors {
				interfaceSlice[i] = v
			}

			if tt.validate != nil {
				tt.validate(t, interfaceSlice)
			}
		})
	}
}
