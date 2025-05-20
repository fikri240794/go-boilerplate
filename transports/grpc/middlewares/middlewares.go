package middlewares

import "google.golang.org/grpc"

type Middlewares struct {
	Recover   *RecoverMiddleware
	Tracer    *TracerMiddleware
	RequestID *RequestIDMiddleware
	Log       *LogMiddleware
	Timeout   *TimeoutMiddleware
}

func (mw *Middlewares) GetUnaryServerInterceptors() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		mw.Recover.Recover,
		mw.Tracer.Start,
		mw.RequestID.Generate,
		mw.Log.Log,
		mw.Timeout.Timeout,
	}
}
