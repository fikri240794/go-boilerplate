package context

import "context"

func SafeCtxValue[T interface{}](ctx context.Context, key interface{}) T {
	var (
		emptyResult *T
		rawCtxValue interface{}
		ctxValue    T
		ok          bool
	)

	emptyResult = new(T)
	rawCtxValue = ctx.Value(key)
	if rawCtxValue == nil {
		return *emptyResult
	}

	ctxValue, ok = rawCtxValue.(T)
	if !ok {
		return *emptyResult
	}

	return ctxValue
}
