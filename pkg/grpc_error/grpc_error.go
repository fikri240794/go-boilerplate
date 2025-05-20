package grpc_error

import (
	"github.com/fikri240794/gocerr"
	"github.com/fikri240794/gostacode"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func FromError(err error) error {
	var (
		customError   gocerr.Error
		isCustomError bool
		grpcCode      codes.Code
		errMessage    string
	)

	if err == nil {
		return nil
	}

	grpcCode = codes.Internal
	customError, isCustomError = gocerr.Parse(err)

	if isCustomError {
		grpcCode = gostacode.GRPCCodeFromHTTPStatusCode(customError.Code)
		errMessage = customError.Message

		if len(customError.ErrorFields) > 0 {
			errMessage = customError.ErrorFields[0].Message
		}
	}

	return status.Error(grpcCode, errMessage)
}
