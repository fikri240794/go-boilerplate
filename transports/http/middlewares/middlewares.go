package middlewares

type Middlewares struct {
	Recover   *RecoverMiddleware
	Tracer    *TracerMiddleware
	RequestID *RequestIDMiddleware
	Log       *LogMiddleware
	Timeout   *TimeoutMiddleware
}
