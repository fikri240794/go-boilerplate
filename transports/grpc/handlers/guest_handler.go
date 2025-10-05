package handlers

import (
	"context"
	"go-boilerplate/internal/models/dtos"
	"go-boilerplate/pkg/constants"
	custom_context "go-boilerplate/pkg/context"
	"go-boilerplate/pkg/grpc_error"
	"go-boilerplate/pkg/protobuf_boilerplate"
	"go-boilerplate/pkg/tracer"
	"go-boilerplate/transports/grpc/models/vms"
	"net/http"

	"github.com/fikri240794/gocerr"
	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (h *ImplementedBoilerplateServer) CreateGuest(ctx context.Context, requestVM *protobuf_boilerplate.CreateGuestRequestVM) (*protobuf_boilerplate.GuestResponseVM, error) {
	var (
		span        trace.Span
		logFields   map[string]interface{}
		requestDTO  *dtos.CreateGuestRequestDTO
		responseDTO *dtos.GuestResponseDTO
		logLevel    zerolog.Level
		responseVM  *protobuf_boilerplate.GuestResponseVM
		err         error
	)

	ctx, span = tracer.Start(ctx, "[ImplementedBoilerplateServer][CreateGuest]")
	defer span.End()

	if requestVM == nil {
		err = grpc_error.FromError(gocerr.New(http.StatusBadRequest, "requestVM is nil"))
		return nil, err
	}

	logFields = map[string]interface{}{
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"requestVM": requestVM,
	}

	requestDTO = vms.CreateGuestRequestVMToDTO(requestVM, uuid.Nil.String())
	logFields["requestDTO"] = requestDTO

	responseDTO, err = h.guestService.Create(ctx, requestDTO)
	if err != nil {
		logLevel = zerolog.WarnLevel
		if gocerr.GetErrorCode(err) >= http.StatusInternalServerError {
			logLevel = zerolog.ErrorLevel
		}

		err = grpc_error.FromError(err)
		log.WithLevel(logLevel).
			Ctx(ctx).
			Err(err).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[ImplementedBoilerplateServer][CreateGuest][Create] failed to create")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[ImplementedBoilerplateServer][CreateGuest][Create] failed to create")
		return nil, err
	}

	responseVM = vms.NewGuestResponseVM(responseDTO)
	return responseVM, nil
}

func (h *ImplementedBoilerplateServer) DeleteGuestByID(ctx context.Context, requestVM *protobuf_boilerplate.DeleteGuestByIDRequestVM) (*emptypb.Empty, error) {
	var (
		span       trace.Span
		logFields  map[string]interface{}
		requestDTO *dtos.DeleteGuestByIDRequestDTO
		logLevel   zerolog.Level
		err        error
	)

	ctx, span = tracer.Start(ctx, "[ImplementedBoilerplateServer][DeleteGuestByID]")
	defer span.End()

	if requestVM == nil {
		err = grpc_error.FromError(gocerr.New(http.StatusBadRequest, "requestVM is nil"))
		return nil, err
	}

	logFields = map[string]interface{}{
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"requestVM": requestVM,
	}

	requestDTO = vms.DeleteGuestByIDRequestVMToDTO(requestVM, uuid.Nil.String())
	logFields["requestDTO"] = requestDTO

	err = h.guestService.DeleteByID(ctx, requestDTO)
	if err != nil {
		logLevel = zerolog.WarnLevel
		if gocerr.GetErrorCode(err) >= fiber.StatusInternalServerError {
			logLevel = zerolog.ErrorLevel
		}

		err = grpc_error.FromError(err)
		log.WithLevel(logLevel).
			Ctx(ctx).
			Err(err).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[ImplementedBoilerplateServer][DeleteGuestByID][DeleteByID] failed to delete by id")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[ImplementedBoilerplateServer][DeleteGuestByID][DeleteByID] failed to delete by id")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *ImplementedBoilerplateServer) FindAllGuest(ctx context.Context, requestVM *protobuf_boilerplate.FindAllGuestRequestVM) (*protobuf_boilerplate.FindAllGuestResponseVM, error) {
	var (
		span        trace.Span
		logFields   map[string]interface{}
		requestDTO  *dtos.FindAllGuestRequestDTO
		responseDTO *dtos.FindAllGuestResponseDTO
		logLevel    zerolog.Level
		responseVM  *protobuf_boilerplate.FindAllGuestResponseVM
		err         error
	)

	ctx, span = tracer.Start(ctx, "[ImplementedBoilerplateServer][FindAllGuest]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"requestVM": requestVM,
	}

	requestDTO = dtos.NewFindAllGuestRequestDTO()
	if requestVM != nil {
		requestDTO = vms.FindAllGuestRequestVMToDTO(requestVM)
	}
	logFields["requestDTO"] = requestDTO

	responseDTO, err = h.guestService.FindAll(ctx, requestDTO)
	if err != nil {
		logLevel = zerolog.WarnLevel
		if gocerr.GetErrorCode(err) >= fiber.StatusInternalServerError {
			logLevel = zerolog.ErrorLevel
		}

		err = grpc_error.FromError(err)
		log.WithLevel(logLevel).
			Ctx(ctx).
			Err(err).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[ImplementedBoilerplateServer][FindAllGuest][FindAll] failed to find all")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[ImplementedBoilerplateServer][FindAllGuest][FindAll] failed to find all")
		return nil, err
	}

	responseVM = vms.NewFindAllGuestResponseVM(responseDTO)
	return responseVM, nil
}

func (h *ImplementedBoilerplateServer) FindGuestByID(ctx context.Context, requestVM *protobuf_boilerplate.FindGuestByIDRequestVM) (*protobuf_boilerplate.GuestResponseVM, error) {
	var (
		span        trace.Span
		logFields   map[string]interface{}
		requestDTO  *dtos.FindGuestByIDRequestDTO
		responseDTO *dtos.GuestResponseDTO
		logLevel    zerolog.Level
		responseVM  *protobuf_boilerplate.GuestResponseVM
		err         error
	)

	ctx, span = tracer.Start(ctx, "[ImplementedBoilerplateServer][FindGuestByID]")
	defer span.End()

	if requestVM == nil {
		err = grpc_error.FromError(gocerr.New(http.StatusBadRequest, "requestVM is nil"))
		return nil, err
	}

	logFields = map[string]interface{}{
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"requestVM": requestVM,
	}

	requestDTO = vms.FindGuestByIDRequestVMToDTO(requestVM)
	logFields["requestDTO"] = requestDTO

	responseDTO, err = h.guestService.FindByID(ctx, requestDTO)
	if err != nil {
		logLevel = zerolog.WarnLevel
		if gocerr.GetErrorCode(err) >= fiber.StatusInternalServerError {
			logLevel = zerolog.ErrorLevel
		}

		err = grpc_error.FromError(err)
		log.WithLevel(logLevel).
			Ctx(ctx).
			Err(err).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[ImplementedBoilerplateServer][FindGuestByID][FindByID] failed to find by id")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[ImplementedBoilerplateServer][FindGuestByID][FindByID] failed to find by id")
		return nil, err
	}

	responseVM = vms.NewGuestResponseVM(responseDTO)
	return responseVM, nil
}

func (h *ImplementedBoilerplateServer) UpdateGuestByID(ctx context.Context, requestVM *protobuf_boilerplate.UpdateGuestByIDRequestVM) (*protobuf_boilerplate.GuestResponseVM, error) {
	var (
		span        trace.Span
		logFields   map[string]interface{}
		requestDTO  *dtos.UpdateGuestByIDRequestDTO
		responseDTO *dtos.GuestResponseDTO
		logLevel    zerolog.Level
		responseVM  *protobuf_boilerplate.GuestResponseVM
		err         error
	)

	ctx, span = tracer.Start(ctx, "[ImplementedBoilerplateServer][UpdateGuestByID]")
	defer span.End()

	if requestVM == nil {
		err = grpc_error.FromError(gocerr.New(http.StatusBadRequest, "requestVM is nil"))
		return nil, err
	}

	logFields = map[string]interface{}{
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"requestVM": requestVM,
	}

	requestDTO = vms.UpdateGuestByIDRequestVMToDTO(requestVM, uuid.Nil.String())
	logFields["requestDTO"] = requestDTO

	responseDTO, err = h.guestService.UpdateByID(ctx, requestDTO)
	if err != nil {
		logLevel = zerolog.WarnLevel
		if gocerr.GetErrorCode(err) >= fiber.StatusInternalServerError {
			logLevel = zerolog.ErrorLevel
		}

		err = grpc_error.FromError(err)
		log.WithLevel(logLevel).
			Ctx(ctx).
			Err(err).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[ImplementedBoilerplateServer][UpdateGuestByID][UpdateByID] failed to update by id")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[ImplementedBoilerplateServer][UpdateGuestByID][UpdateByID] failed to update by id")
		return nil, err
	}

	responseVM = vms.NewGuestResponseVM(responseDTO)
	return responseVM, nil
}
