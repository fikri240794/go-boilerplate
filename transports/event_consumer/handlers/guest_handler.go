package handlers

import (
	"context"
	"go-boilerplate/internal/models/dtos"
	"go-boilerplate/internal/services"
	"go-boilerplate/pkg/constants"
	custom_context "go-boilerplate/pkg/context"
	"go-boilerplate/pkg/tracer"
	"go-boilerplate/transports/event_consumer/models/vms"
	"net/http"

	"github.com/fikri240794/gocerr"
	"github.com/goccy/go-json"
	"github.com/nsqio/go-nsq"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

type GuestHandler struct {
	guestService services.IGuestService
}

func NewGuestHandler(guestService services.IGuestService) *GuestHandler {
	return &GuestHandler{
		guestService: guestService,
	}
}

func (h *GuestHandler) HandleCreated(ctx context.Context, m *nsq.Message) error {
	var (
		span       trace.Span
		logFields  map[string]interface{}
		requestVM  *vms.EventRequestVM[vms.GuestEventRequestVM]
		requestDTO *dtos.GuestEventRequestDTO
		err        error
	)

	ctx, span = tracer.Start(ctx, "[GuestHandler][HandleCreated]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid":   custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID),
		"messageBody": string(m.Body),
	}

	log.Debug().
		Ctx(ctx).
		Fields(logFields).
		Msg("[GuestHandler][HandleCreated] message received")

	requestVM = &vms.EventRequestVM[vms.GuestEventRequestVM]{}
	err = json.Unmarshal(m.Body, requestVM)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[GuestHandler][HandleCreated][Unmarshal] failed to parse message body")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestHandler][HandleCreated][Unmarshal] failed to parse message body")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}
	logFields["requestVM"] = requestVM

	if requestVM.Message == nil {
		err = gocerr.New(http.StatusInternalServerError, "message is nil")
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[GuestHandler][HandleCreated] message is nil")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestHandler][HandleCreated] message is nil")
		return err
	}

	requestDTO = requestVM.Message.ToDTO()

	_, err = h.guestService.ProcessEvent(ctx, requestDTO)
	if err != nil {
		logFields["requestDTO"] = requestDTO
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[GuestHandler][HandleCreated][ProcessEvent] failed to process event")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestHandler][HandleCreated][ProcessEvent] failed to process event")
		return err
	}

	return nil
}

func (h *GuestHandler) HandleDeleted(ctx context.Context, m *nsq.Message) error {
	var (
		span       trace.Span
		logFields  map[string]interface{}
		requestVM  *vms.EventRequestVM[vms.GuestEventRequestVM]
		requestDTO *dtos.GuestEventRequestDTO
		err        error
	)

	ctx, span = tracer.Start(ctx, "[GuestHandler][HandleDeleted]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid":   custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID),
		"messageBody": string(m.Body),
	}

	log.Debug().
		Ctx(ctx).
		Fields(logFields).
		Msg("[GuestHandler][HandleDeleted] message received")

	requestVM = &vms.EventRequestVM[vms.GuestEventRequestVM]{}
	err = json.Unmarshal(m.Body, requestVM)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[GuestHandler][HandleDeleted][Unmarshal] failed to parse message body")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestHandler][HandleDeleted][Unmarshal] failed to parse message body")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}
	logFields["requestVM"] = requestVM

	if requestVM.Message == nil {
		err = gocerr.New(http.StatusInternalServerError, "message is nil")
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[GuestHandler][HandleDeleted] message is nil")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestHandler][HandleDeleted] message is nil")
		return err
	}

	requestDTO = requestVM.Message.ToDTO()

	_, err = h.guestService.ProcessEvent(ctx, requestDTO)
	if err != nil {
		logFields["requestDTO"] = requestDTO
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[GuestHandler][HandleDeleted][ProcessEvent] failed to process event")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestHandler][HandleDeleted][ProcessEvent] failed to process event")
		return err
	}

	return nil
}

func (h *GuestHandler) HandleUpdated(ctx context.Context, m *nsq.Message) error {
	var (
		logFields  map[string]interface{}
		requestVM  *vms.EventRequestVM[vms.GuestEventRequestVM]
		requestDTO *dtos.GuestEventRequestDTO
		err        error
	)

	logFields = map[string]interface{}{
		"requestid":   custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID),
		"messageBody": string(m.Body),
	}

	log.Debug().
		Ctx(ctx).
		Fields(logFields).
		Msg("[GuestHandler][HandleUpdated] message received")

	requestVM = &vms.EventRequestVM[vms.GuestEventRequestVM]{}
	err = json.Unmarshal(m.Body, requestVM)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[GuestHandler][HandleUpdated][Unmarshal] failed to parse message body")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestHandler][HandleUpdated][Unmarshal] failed to parse message body")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}
	logFields["requestVM"] = requestVM

	if requestVM.Message == nil {
		err = gocerr.New(http.StatusInternalServerError, "message is nil")
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[GuestHandler][HandleUpdated] message is nil")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestHandler][HandleUpdated] message is nil")
		return err
	}

	requestDTO = requestVM.Message.ToDTO()

	_, err = h.guestService.ProcessEvent(ctx, requestDTO)
	if err != nil {
		logFields["requestDTO"] = requestDTO
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[GuestHandler][HandleUpdated][ProcessEvent] failed to process event")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestHandler][HandleUpdated][ProcessEvent] failed to process event")
		return err
	}

	return nil
}
