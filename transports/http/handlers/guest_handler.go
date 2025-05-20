package handlers

import (
	"context"
	"go-boilerplate/internal/models/dtos"
	"go-boilerplate/internal/services"
	"go-boilerplate/pkg/constants"
	custom_context "go-boilerplate/pkg/context"
	"go-boilerplate/pkg/tracer"
	"go-boilerplate/transports/http/models/vms"

	"github.com/fikri240794/gocerr"
	"github.com/fikri240794/gores"
	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog"
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

func (h *GuestHandler) SetupRoutes(server *fiber.App) {
	server.Route("/guests", func(api fiber.Router) {
		api.Post("/", h.Create)
		api.Delete("/:id", h.DeleteByID)
		api.Get("/", h.FindAll)
		api.Get("/:id", h.FindByID)
		api.Put("/:id", h.UpdateByID)
	})
}

// @Summary	Create Guest
// @Description	Create Guest
// @Tags	guest
// @Accept	application/json
// @Produce	application/json
// @Param	CreateGuestRequestVM	body	vms.CreateGuestRequestVM	true	"CreateGuestRequestVM"
// @Success	201	{object}	gores.ResponseVM[vms.GuestResponseVM]
// @Failure	400	{object}	gores.ResponseVM[vms.GuestResponseVM]
// @Failure	500	{object}	gores.ResponseVM[vms.GuestResponseVM]
// @Router	/guests	[post]
func (h *GuestHandler) Create(c *fiber.Ctx) error {
	var (
		ctx         context.Context
		span        trace.Span
		logFields   map[string]interface{}
		requestVM   *vms.CreateGuestRequestVM
		requestDTO  *dtos.CreateGuestRequestDTO
		responseDTO *dtos.GuestResponseDTO
		logLevel    zerolog.Level
		responseVM  *gores.ResponseVM[*vms.GuestResponseVM]
		err         error
	)

	ctx = c.UserContext()

	ctx, span = tracer.Start(ctx, "[GuestHandler][Create]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid": custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID),
	}

	requestVM = &vms.CreateGuestRequestVM{}
	err = c.BodyParser(requestVM)
	if err != nil {
		log.Warn().
			Ctx(ctx).
			Err(err).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[GuestHandler][Create][BodyParser] failed to parse request body")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestHandler][Create][BodyParser] failed to parse request body")
		err = gocerr.New(fiber.StatusBadRequest, err.Error())
		responseVM = gores.NewResponseVM[*vms.GuestResponseVM]().
			SetErrorFromError(err)
		return c.Status(responseVM.Code).
			JSON(responseVM)
	}
	logFields["requestVM"] = requestVM

	requestDTO = requestVM.ToDTO(uuid.Nil.String())
	logFields["requestDTO"] = requestDTO

	responseDTO, err = h.guestService.Create(ctx, requestDTO)
	if err != nil {
		logLevel = zerolog.WarnLevel
		if gocerr.GetErrorCode(err) >= fiber.StatusInternalServerError {
			logLevel = zerolog.ErrorLevel
		}

		log.WithLevel(logLevel).
			Ctx(ctx).
			Err(err).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[GuestHandler][Create][Create] failed to create")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestHandler][Create][Create] failed to create")
		responseVM = gores.NewResponseVM[*vms.GuestResponseVM]().
			SetErrorFromError(err)
		return c.Status(responseVM.Code).
			JSON(responseVM)
	}

	responseVM = gores.NewResponseVM[*vms.GuestResponseVM]().
		SetCode(fiber.StatusCreated).
		SetData(vms.NewGuestResponseVM(responseDTO))

	return c.Status(responseVM.Code).
		JSON(responseVM)
}

// @Summary	Delete Guest by ID
// @Description	Delete Guest by ID
// @Tags	guest
// @Produce	application/json
// @Param	id	path	string	true	"id"	example(01932293-d710-7f55-a9f6-66e6248ae72f)
// @Success	200	{object}	gores.ResponseVM[bool]
// @Failure	400	{object}	gores.ResponseVM[bool]
// @Failure	404	{object}	gores.ResponseVM[bool]
// @Failure	500	{object}	gores.ResponseVM[bool]
// @Router	/guests/{id}	[delete]
func (h *GuestHandler) DeleteByID(c *fiber.Ctx) error {
	var (
		ctx        context.Context
		span       trace.Span
		logFields  map[string]interface{}
		requestVM  *vms.DeleteGuestByIDRequestVM
		requestDTO *dtos.DeleteGuestByIDRequestDTO
		logLevel   zerolog.Level
		responseVM *gores.ResponseVM[bool]
		err        error
	)

	ctx = c.UserContext()

	ctx, span = tracer.Start(ctx, "[GuestHandler][DeleteByID]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid": custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID),
	}

	requestVM = &vms.DeleteGuestByIDRequestVM{}
	err = c.ParamsParser(requestVM)
	if err != nil {
		log.Warn().
			Ctx(ctx).
			Err(err).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[GuestHandler][DeleteByID][ParamsParser] failed to parse request params")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestHandler][DeleteByID][ParamsParser] failed to parse request params")
		err = gocerr.New(fiber.StatusBadRequest, err.Error())
		responseVM = gores.NewResponseVM[bool]().
			SetErrorFromError(err)
		return c.Status(responseVM.Code).
			JSON(responseVM)
	}
	logFields["requestVM"] = requestVM

	requestDTO = requestVM.ToDTO(uuid.Nil.String())
	logFields["requestDTO"] = requestDTO

	err = h.guestService.DeleteByID(ctx, requestDTO)
	if err != nil {
		logLevel = zerolog.WarnLevel
		if gocerr.GetErrorCode(err) >= fiber.StatusInternalServerError {
			logLevel = zerolog.ErrorLevel
		}

		log.WithLevel(logLevel).
			Ctx(ctx).
			Err(err).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[GuestHandler][DeleteByID][DeleteByID] failed to delete by id")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestHandler][DeleteByID][DeleteByID] failed to delete by id")
		responseVM = gores.NewResponseVM[bool]().
			SetErrorFromError(err)
		return c.Status(responseVM.Code).
			JSON(responseVM)
	}

	responseVM = gores.NewResponseVM[bool]().
		SetCode(fiber.StatusOK).
		SetData(true)

	return c.Status(responseVM.Code).
		JSON(responseVM)
}

// @Summary	Find All Guest
// @Description	Find All Guest
// @Tags	guest
// @Produce	application/json
// @Param	keyword	query	string	false	"name or address"	example(John Snow or 123 Main Street)
// @Param	sorts	query	string	false	"sorts"	example(name.asc,address.desc)
// @Param	take	query	number	true	"take"	example(10)	minimum(1)
// @Param	skip	query	number	false	"skip"	example(0)	minimum(0)
// @Success	200	{object}	gores.ResponseVM[vms.FindAllGuestResponseVM]
// @Failure	400	{object}	gores.ResponseVM[vms.FindAllGuestResponseVM]
// @Failure	500	{object}	gores.ResponseVM[vms.FindAllGuestResponseVM]
// @Router	/guests	[get]
func (h *GuestHandler) FindAll(c *fiber.Ctx) error {
	var (
		ctx         context.Context
		span        trace.Span
		logFields   map[string]interface{}
		requestVM   *vms.FindAllGuestRequestVM
		requestDTO  *dtos.FindAllGuestRequestDTO
		responseDTO *dtos.FindAllGuestResponseDTO
		logLevel    zerolog.Level
		responseVM  *gores.ResponseVM[*vms.FindAllGuestResponseVM]
		err         error
	)

	ctx = c.UserContext()

	ctx, span = tracer.Start(ctx, "[GuestHandler][FindAll]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid": custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID),
	}

	requestVM = &vms.FindAllGuestRequestVM{}
	err = c.QueryParser(requestVM)
	if err != nil {
		log.Warn().
			Ctx(ctx).
			Err(err).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[GuestHandler][FindAll][QueryParser] failed to parse request query")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestHandler][FindAll][QueryParser] failed to parse request query")
		err = gocerr.New(fiber.StatusBadRequest, err.Error())
		responseVM = gores.NewResponseVM[*vms.FindAllGuestResponseVM]().
			SetErrorFromError(err)
		return c.Status(responseVM.Code).
			JSON(responseVM)
	}
	logFields["requestVM"] = requestVM

	requestDTO = requestVM.ToDTO()
	logFields["requestDTO"] = requestDTO

	responseDTO, err = h.guestService.FindAll(ctx, requestDTO)
	if err != nil {
		logLevel = zerolog.WarnLevel
		if gocerr.GetErrorCode(err) >= fiber.StatusInternalServerError {
			logLevel = zerolog.ErrorLevel
		}

		log.WithLevel(logLevel).
			Ctx(ctx).
			Err(err).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[GuestHandler][FindAll][FindAll] failed to find all")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestHandler][FindAll][FindAll] failed to find all")
		responseVM = gores.NewResponseVM[*vms.FindAllGuestResponseVM]().
			SetErrorFromError(err)
		return c.Status(responseVM.Code).
			JSON(responseVM)
	}

	responseVM = gores.NewResponseVM[*vms.FindAllGuestResponseVM]().
		SetCode(fiber.StatusOK).
		SetData(vms.NewFindAllGuestResponseVM(responseDTO))

	return c.Status(responseVM.Code).
		JSON(responseVM)
}

// @Summary	Find Guest by ID
// @Description	Find Guest by ID
// @Tags	guest
// @Produce	application/json
// @Param	id	path	string	true	"id"  example(01932293-d710-7f55-a9f6-66e6248ae72f)
// @Success	200	{object}	gores.ResponseVM[vms.GuestResponseVM]
// @Failure	400	{object}	gores.ResponseVM[vms.GuestResponseVM]
// @Failure	404	{object}	gores.ResponseVM[vms.GuestResponseVM]
// @Failure	500	{object}	gores.ResponseVM[vms.GuestResponseVM]
// @Router	/guests/{id}	[get]
func (h *GuestHandler) FindByID(c *fiber.Ctx) error {
	var (
		ctx         context.Context
		span        trace.Span
		logFields   map[string]interface{}
		requestVM   *vms.FindGuestByIDRequestVM
		requestDTO  *dtos.FindGuestByIDRequestDTO
		responseDTO *dtos.GuestResponseDTO
		logLevel    zerolog.Level
		responseVM  *gores.ResponseVM[*vms.GuestResponseVM]
		err         error
	)

	ctx = c.UserContext()

	ctx, span = tracer.Start(ctx, "[GuestHandler][FindByID]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid": custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID),
	}

	requestVM = &vms.FindGuestByIDRequestVM{}
	err = c.ParamsParser(requestVM)
	if err != nil {
		log.Warn().
			Ctx(ctx).
			Err(err).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[GuestHandler][FindByID][ParamsParser] failed to parse request params")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestHandler][FindByID][ParamsParser] failed to parse request params")
		err = gocerr.New(fiber.StatusBadRequest, err.Error())
		responseVM = gores.NewResponseVM[*vms.GuestResponseVM]().
			SetErrorFromError(err)
		return c.Status(responseVM.Code).
			JSON(responseVM)
	}
	logFields["requestVM"] = requestVM

	requestDTO = requestVM.ToDTO()
	logFields["requestDTO"] = requestDTO

	responseDTO, err = h.guestService.FindByID(ctx, requestDTO)
	if err != nil {
		logLevel = zerolog.WarnLevel
		if gocerr.GetErrorCode(err) >= fiber.StatusInternalServerError {
			logLevel = zerolog.ErrorLevel
		}

		log.WithLevel(logLevel).
			Ctx(ctx).
			Err(err).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[GuestHandler][FindByID][FindByID] failed to find by id")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestHandler][FindByID][FindByID] failed to find by id")
		responseVM = gores.NewResponseVM[*vms.GuestResponseVM]().
			SetErrorFromError(err)
		return c.Status(responseVM.Code).
			JSON(responseVM)
	}

	responseVM = gores.NewResponseVM[*vms.GuestResponseVM]().
		SetCode(fiber.StatusOK).
		SetData(vms.NewGuestResponseVM(responseDTO))

	return c.Status(responseVM.Code).
		JSON(responseVM)
}

// @Summary	Update Guest by ID
// @Description	Update Guest by ID
// @Tags	guest
// @Produce	application/json
// @Param	id	path	string	true	"id"  example(01932293-d710-7f55-a9f6-66e6248ae72f)
// @Param	UpdateGuestByIDRequestVM	body	vms.UpdateGuestByIDRequestVM	true	"UpdateGuestByIDRequestVM"
// @Success	200	{object}	gores.ResponseVM[vms.GuestResponseVM]
// @Failure	400	{object}	gores.ResponseVM[vms.GuestResponseVM]
// @Failure	404	{object}	gores.ResponseVM[vms.GuestResponseVM]
// @Failure	500	{object}	gores.ResponseVM[vms.GuestResponseVM]
// @Router	/guests/{id}	[put]
func (h *GuestHandler) UpdateByID(c *fiber.Ctx) error {
	var (
		ctx         context.Context
		span        trace.Span
		logFields   map[string]interface{}
		requestVM   *vms.UpdateGuestByIDRequestVM
		requestDTO  *dtos.UpdateGuestByIDRequestDTO
		responseDTO *dtos.GuestResponseDTO
		logLevel    zerolog.Level
		responseVM  *gores.ResponseVM[*vms.GuestResponseVM]
		err         error
	)

	ctx = c.UserContext()

	ctx, span = tracer.Start(ctx, "[GuestHandler][UpdateByID]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid": custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID),
	}

	requestVM = &vms.UpdateGuestByIDRequestVM{}
	err = c.ParamsParser(requestVM)
	if err != nil {
		log.Warn().
			Ctx(ctx).
			Err(err).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[GuestHandler][UpdateByID][ParamsParser] failed to parse request params")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestHandler][UpdateByID][ParamsParser] failed to parse request params")
		err = gocerr.New(fiber.StatusBadRequest, err.Error())
		responseVM = gores.NewResponseVM[*vms.GuestResponseVM]().
			SetErrorFromError(err)
		return c.Status(responseVM.Code).
			JSON(responseVM)
	}
	logFields["requestVM"] = requestVM

	err = c.BodyParser(requestVM)
	if err != nil {
		log.Warn().
			Ctx(ctx).
			Err(err).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[GuestHandler][UpdateByID][BodyParser] failed to parse request body")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestHandler][UpdateByID][BodyParser] failed to parse request body")
		err = gocerr.New(fiber.StatusBadRequest, err.Error())
		responseVM = gores.NewResponseVM[*vms.GuestResponseVM]().
			SetErrorFromError(err)
		return c.Status(responseVM.Code).
			JSON(responseVM)
	}
	logFields["requestVM"] = requestVM

	requestDTO = requestVM.ToDTO(uuid.Nil.String())
	logFields["requestDTO"] = requestDTO

	responseDTO, err = h.guestService.UpdateByID(ctx, requestDTO)
	if err != nil {
		logLevel = zerolog.WarnLevel
		if gocerr.GetErrorCode(err) >= fiber.StatusInternalServerError {
			logLevel = zerolog.ErrorLevel
		}

		log.WithLevel(logLevel).
			Ctx(ctx).
			Err(err).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[GuestHandler][UpdateByID][UpdateByID] failed to update by id")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestHandler][UpdateByID][UpdateByID] failed to update by id")
		responseVM = gores.NewResponseVM[*vms.GuestResponseVM]().
			SetErrorFromError(err)
		return c.Status(responseVM.Code).
			JSON(responseVM)
	}

	responseVM = gores.NewResponseVM[*vms.GuestResponseVM]().
		SetCode(fiber.StatusOK).
		SetData(vms.NewGuestResponseVM(responseDTO))

	return c.Status(responseVM.Code).
		JSON(responseVM)
}
