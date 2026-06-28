package services

import (
	"context"
	"fmt"
	"go-boilerplate/configs"
	"go-boilerplate/internal/models/dtos"
	"go-boilerplate/internal/models/entities"
	"go-boilerplate/internal/repositories"
	"go-boilerplate/pkg/tracer"
	"net/http"
	"regexp"
	"strings"

	"github.com/fikri240794/gocerr"
	"github.com/fikri240794/goqube"
	"github.com/fikri240794/gotask"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

//mockery:generate: true
//mockery:structname: GuestServiceMock
//mockery:filename: guest_service_mock.go
//mockery:output: internal/services/mocks/
type IGuestService interface {
	BulkCreate(ctx context.Context, requestDTO *dtos.BulkCreateGuestsRequestDTO) (*dtos.BulkCreateGuestsResponseDTO, error)
	BulkDelete(ctx context.Context, requestDTO *dtos.BulkDeleteGuestsRequestDTO) error
	BulkUpdate(ctx context.Context, requestDTO *dtos.BulkUpdateGuestsRequestDTO) (*dtos.BulkUpdateGuestsResponseDTO, error)
	Create(ctx context.Context, requestDTO *dtos.CreateGuestRequestDTO) (*dtos.GuestResponseDTO, error)
	DeleteByID(ctx context.Context, requestDTO *dtos.DeleteGuestByIDRequestDTO) error
	FindAll(ctx context.Context, requestDTO *dtos.FindAllGuestRequestDTO) (*dtos.FindAllGuestResponseDTO, error)
	FindByID(ctx context.Context, requestDTO *dtos.FindGuestByIDRequestDTO) (*dtos.GuestResponseDTO, error)
	UpdateByID(ctx context.Context, requestDTO *dtos.UpdateGuestByIDRequestDTO) (*dtos.GuestResponseDTO, error)
	ProcessEvent(ctx context.Context, requestDTO *dtos.GuestEventRequestDTO) (*dtos.GuestEventResponseDTO, error)
}

type GuestService struct {
	cfg                          *configs.Config
	guestRepository              repositories.IGuestRepository
	guestCacheRepository         repositories.IGuestCacheRepository
	guestEventProducerRepository repositories.IGuestEventProducerRepository
	webhookSiteRepository        repositories.IWebhookSiteRepository
}

func NewGuestService(
	cfg *configs.Config,
	guestRepository repositories.IGuestRepository,
	guestCacheRepository repositories.IGuestCacheRepository,
	guestEventProducerRepository repositories.IGuestEventProducerRepository,
	webhookSiteRepository repositories.IWebhookSiteRepository,
) *GuestService {
	return &GuestService{
		cfg:                          cfg,
		guestRepository:              guestRepository,
		guestCacheRepository:         guestCacheRepository,
		guestEventProducerRepository: guestEventProducerRepository,
		webhookSiteRepository:        webhookSiteRepository,
	}
}

func (s *GuestService) deleteEntityCaches(ctx context.Context) error {
	var (
		span      trace.Span
		logFields map[string]interface{}
		pattern   string
		keys      []string
		err       error
	)

	ctx, span = tracer.Start(ctx, "[GuestService][deleteEntityCaches]")
	defer span.End()

	logFields = map[string]interface{}{}

	pattern = fmt.Sprintf(s.cfg.Guest.Cache.Keyf, "*")
	logFields["pattern"] = pattern

	keys, err = s.guestCacheRepository.Keys(ctx, pattern)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[GuestService][deleteEntityCaches][Keys] failed to get cache keys")
		return err
	}

	if len(keys) <= 0 {
		return nil
	}

	logFields["keys"] = keys

	err = s.guestCacheRepository.Delete(ctx, keys...)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[GuestService][deleteEntityCaches][Delete] failed to delete caches")
		return err
	}

	return nil
}

func (s *GuestService) withTransaction(
	ctx context.Context,
	logFields map[string]interface{},
	fnName string,
	fn func(repositories.IBoilerplateDatabaseTransaction) error,
) error {
	var (
		tx          repositories.IBoilerplateDatabaseTransaction
		errRollback error
		err         error
	)

	tx, err = s.guestRepository.BeginTransaction(ctx)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg(fmt.Sprintf("[GuestService][%s][BeginTransaction] failed to begin transaction", fnName))
		return err
	}

	defer func() {
		if err != nil {
			errRollback = tx.Rollback()
			if errRollback != nil {
				log.Err(errRollback).
					Ctx(ctx).
					Fields(logFields).
					Msg(fmt.Sprintf("[GuestService][%s][Rollback] failed to rollback transaction", fnName))
			}
		}
	}()

	err = fn(tx)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg(fmt.Sprintf("[GuestService][%s][WithTransaction] failed to execute operation", fnName))
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg(fmt.Sprintf("[GuestService][%s][Commit] failed to commit transaction", fnName))
		return err
	}

	return nil
}

func (s *GuestService) buildActiveEntityFilterByIDs(ids ...string) *goqube.Filter {
	var (
		operator goqube.Operator
		value    interface{}
	)

	if len(ids) == 1 {
		operator = goqube.OperatorEqual
		value = ids[0]
	} else {
		operator = goqube.OperatorIn
		value = ids
	}

	return &goqube.Filter{
		Logic: goqube.LogicAnd,
		Filters: []goqube.Filter{
			{
				Field:    goqube.Field{Column: entities.GuestEntityDatabaseFieldID},
				Operator: operator,
				Value:    goqube.FilterValue{Value: value},
			},
			{
				Field:    goqube.Field{Column: entities.GuestEntityDatabaseFieldDeletedAt},
				Operator: goqube.OperatorIsNull,
				Value:    goqube.FilterValue{Value: nil},
			},
		},
	}
}

func (s *GuestService) tryDeleteEntityCaches(ctx context.Context, logFields map[string]interface{}, fnName string) {
	var err error

	err = s.deleteEntityCaches(ctx)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg(fmt.Sprintf("[GuestService][%s][deleteEntityCaches] failed to delete caches", fnName))
	}
}

func (s *GuestService) publishEvent(
	ctx context.Context,
	logFields map[string]interface{},
	enable bool,
	topic string,
	fnName string,
	entities_ ...entities.GuestEntity,
) {
	var (
		eventEntity   *entities.EventEntity[entities.GuestEventEntity]
		eventEntities []entities.GuestEventEntity
		err           error
	)

	if !enable {
		return
	}

	logFields["eventTopic"] = topic

	if len(entities_) == 1 {
		eventEntity = entities.NewEventEntity(topic, entities.NewGuestEventEntity(&entities_[0]))
		logFields["eventEntity"] = eventEntity

		err = s.guestEventProducerRepository.Publish(ctx, topic, eventEntity)
		if err != nil {
			log.Err(err).
				Ctx(ctx).
				Fields(logFields).
				Msg(fmt.Sprintf("[GuestService][%s][Publish] failed to publish message", fnName))
		}
		return
	}

	for i := range entities_ {
		eventEntities = append(eventEntities, *entities.NewGuestEventEntity(&entities_[i]))
	}
	logFields["eventEntities"] = eventEntities

	err = s.guestEventProducerRepository.PublishBulk(
		ctx,
		topic,
		entities.NewEventEntity(topic, &eventEntities),
	)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg(fmt.Sprintf("[GuestService][%s][PublishBulk] failed to publish message", fnName))
	}
}

func (s *GuestService) Create(ctx context.Context, requestDTO *dtos.CreateGuestRequestDTO) (*dtos.GuestResponseDTO, error) {
	var (
		span        trace.Span
		logFields   map[string]interface{}
		entity      *entities.GuestEntity
		responseDTO *dtos.GuestResponseDTO
		err         error
	)

	ctx, span = tracer.Start(ctx, "[GuestService][Create]")
	defer span.End()

	if requestDTO == nil {
		return nil, gocerr.New(http.StatusBadRequest, "requestDTO is nil")
	}

	logFields = map[string]interface{}{
		"requestDTO": requestDTO,
	}

	err = requestDTO.Validate()
	if err != nil {
		log.Warn().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestService][Create][Validate] failed to validate dto")
		return nil, err
	}

	entity = requestDTO.ToEntity()
	logFields["entity"] = entity

	err = s.withTransaction(ctx, logFields, "Create", func(tx repositories.IBoilerplateDatabaseTransaction) error {
		return s.guestRepository.WithTransaction(tx).Create(ctx, entity)
	})
	if err != nil {
		return nil, err
	}

	responseDTO = dtos.NewGuestResponseDTO(entity)
	logFields["responseDTO"] = responseDTO

	s.tryDeleteEntityCaches(ctx, logFields, "Create")
	s.publishEvent(ctx, logFields, s.cfg.Guest.Event.Created.Enable, s.cfg.Guest.Event.Created.Topic, "Create", *entity)

	return responseDTO, nil
}

func (s *GuestService) DeleteByID(ctx context.Context, requestDTO *dtos.DeleteGuestByIDRequestDTO) error {
	var (
		span      trace.Span
		logFields map[string]interface{}
		filter    *goqube.Filter
		entity    *entities.GuestEntity
		logLevel  zerolog.Level
		err       error
	)

	ctx, span = tracer.Start(ctx, "[GuestService][DeleteByID]")
	defer span.End()

	if requestDTO == nil {
		return gocerr.New(http.StatusBadRequest, "requestDTO is nil")
	}

	logFields = map[string]interface{}{
		"requestDTO": requestDTO,
	}

	err = requestDTO.Validate()
	if err != nil {
		log.Warn().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestService][DeleteByID][Validate] failed to validate dto")
		return err
	}

	filter = s.buildActiveEntityFilterByIDs(requestDTO.ID)
	logFields["filter"] = filter

	entity, err = s.guestRepository.FindOne(ctx, filter, nil, false)
	if err != nil {
		logLevel = zerolog.WarnLevel
		if gocerr.GetErrorCode(err) >= http.StatusInternalServerError {
			logLevel = zerolog.ErrorLevel
		}

		log.WithLevel(logLevel).
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestService][DeleteByID][FindOne] failed to find entity")
		return err
	}

	entity = entity.MarkAsDeleted(requestDTO.DeletedBy)
	logFields["entity"] = entity

	err = s.withTransaction(ctx, logFields, "DeleteByID", func(tx repositories.IBoilerplateDatabaseTransaction) error {
		return s.guestRepository.WithTransaction(tx).Update(ctx, entity, filter)
	})
	if err != nil {
		return err
	}

	s.tryDeleteEntityCaches(ctx, logFields, "DeleteByID")
	s.publishEvent(ctx, logFields, s.cfg.Guest.Event.Deleted.Enable, s.cfg.Guest.Event.Deleted.Topic, "DeleteByID", *entity)

	return nil
}

func (s *GuestService) getListEntityCache(
	ctx context.Context,
	listEntityCacheKey string,
) ([]entities.GuestEntity, error) {
	var (
		span       trace.Span
		logFields  map[string]interface{}
		listEntity []entities.GuestEntity
		err        error
	)

	ctx, span = tracer.Start(ctx, "[GuestService][getListEntityCache]")
	defer span.End()

	logFields = map[string]interface{}{
		"listEntityCacheKey": listEntityCacheKey,
	}

	listEntity, err = s.guestCacheRepository.GetList(ctx, listEntityCacheKey)
	if err != nil {
		if gocerr.GetErrorCode(err) >= http.StatusInternalServerError {
			log.Err(err).
				Ctx(ctx).
				Fields(logFields).
				Msg("[GuestService][getListEntityCache][GetList] failed to get list entity cache")
		}
		return nil, err
	}
	logFields["listEntity"] = listEntity

	return listEntity, nil
}

func (s *GuestService) setListEntityCache(
	ctx context.Context,
	listEntityCacheKey string,
	listEntity []entities.GuestEntity,
) error {
	var (
		span      trace.Span
		logFields map[string]interface{}
		err       error
	)

	ctx, span = tracer.Start(ctx, "[GuestService][setListEntityCache]")
	defer span.End()

	logFields = map[string]interface{}{
		"listEntityCacheKey": listEntityCacheKey,
		"listEntity":         listEntity,
		"expiration":         s.cfg.Guest.Cache.Duration,
	}

	err = s.guestCacheRepository.SetList(
		ctx,
		listEntityCacheKey,
		listEntity,
		s.cfg.Guest.Cache.Duration,
	)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[GuestService][setListEntityCache][SetList] failed to set list entity cache")
		return err
	}

	return nil
}

func (s *GuestService) findListEntity(
	ctx context.Context,
	listEntityCacheKey string,
	filter *goqube.Filter,
	sorts []goqube.Sort,
	take uint64,
	skip uint64,
) ([]entities.GuestEntity, error) {
	var (
		span       trace.Span
		logFields  map[string]interface{}
		listEntity []entities.GuestEntity
		err        error
	)

	ctx, span = tracer.Start(ctx, "[GuestService][findListEntity]")
	defer span.End()

	logFields = map[string]interface{}{
		"listEntityCacheKey": listEntityCacheKey,
		"filter":             filter,
		"sorts":              sorts,
		"take":               take,
		"skip":               skip,
	}

	if s.cfg.Guest.Cache.Enable {
		listEntity, err = s.getListEntityCache(ctx, listEntityCacheKey)
		if err != nil {
			if gocerr.GetErrorCode(err) >= http.StatusInternalServerError {
				log.Err(err).
					Ctx(ctx).
					Fields(logFields).
					Msg("[GuestService][findListEntity][getListEntityCache] failed to find list entity cache")
			}
			err = nil
		}

		if len(listEntity) > 0 {
			return listEntity, nil
		}
	}

	listEntity, err = s.guestRepository.FindAll(
		ctx,
		filter,
		sorts,
		take,
		skip,
		false,
	)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[GuestService][findListEntity][FindAll] failed to find list entity")
		return nil, err
	}
	logFields["listEntity"] = listEntity

	if s.cfg.Guest.Cache.Enable && len(listEntity) > 0 {
		err = s.setListEntityCache(
			ctx,
			listEntityCacheKey,
			listEntity,
		)
		if err != nil {
			log.Err(err).
				Ctx(ctx).
				Fields(logFields).
				Msg("[GuestService][findListEntity][setListEntityCache] failed to set list entity cache")
			err = nil
		}
	}

	return listEntity, nil
}

func (s *GuestService) getCountEntitiesCache(
	ctx context.Context,
	entitiesCountCacheKey string,
) (uint64, error) {
	var (
		span          trace.Span
		logFields     map[string]interface{}
		entitiesCount uint64
		err           error
	)

	ctx, span = tracer.Start(ctx, "[GuestService][getCountEntitiesCache]")
	defer span.End()

	logFields = map[string]interface{}{
		"entitiesCountCacheKey": entitiesCountCacheKey,
	}

	entitiesCount, err = s.guestCacheRepository.GetCount(ctx, entitiesCountCacheKey)
	if err != nil {
		if gocerr.GetErrorCode(err) >= http.StatusInternalServerError {
			log.Err(err).
				Ctx(ctx).
				Fields(logFields).
				Msg("[GuestService][getCountEntitiesCache][GetCount] failed to get entities count cache")
		}
		return 0, err
	}

	return entitiesCount, nil
}

func (s *GuestService) setEntitiesCountCache(
	ctx context.Context,
	entitiesCountCacheKey string,
	entitiesCount uint64,
) error {
	var (
		span      trace.Span
		logFields map[string]interface{}
		err       error
	)

	ctx, span = tracer.Start(ctx, "[GuestService][setEntitiesCountCache]")
	defer span.End()

	logFields = map[string]interface{}{
		"entitiesCountCacheKey": entitiesCountCacheKey,
		"entitiesCount":         entitiesCount,
		"expiration":            s.cfg.Guest.Cache.Duration,
	}

	err = s.guestCacheRepository.SetCount(
		ctx,
		entitiesCountCacheKey,
		entitiesCount,
		s.cfg.Guest.Cache.Duration,
	)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[GuestService][setEntitiesCountCache][SetCount] failed to set entities count cache")
		return err
	}

	return nil
}

func (s *GuestService) countEntities(
	ctx context.Context,
	entitiesCountCacheKey string,
	filter *goqube.Filter,
) (uint64, error) {
	var (
		span          trace.Span
		logFields     map[string]interface{}
		entitiesCount uint64
		err           error
	)

	ctx, span = tracer.Start(ctx, "[GuestService][countEntities]")
	defer span.End()

	logFields = map[string]interface{}{
		"entitiesCountCacheKey": entitiesCountCacheKey,
		"filter":                filter,
	}

	if s.cfg.Guest.Cache.Enable {
		entitiesCount, err = s.getCountEntitiesCache(ctx, entitiesCountCacheKey)
		if err != nil {
			if gocerr.GetErrorCode(err) >= http.StatusInternalServerError {
				log.Err(err).
					Ctx(ctx).
					Fields(logFields).
					Msg("[GuestService][countEntities][getCountEntitiesCache] failed to get entities count cache")
			}

			err = nil
		}

		if entitiesCount > 0 {
			return entitiesCount, nil
		}
	}

	entitiesCount, err = s.guestRepository.Count(
		ctx,
		filter,
		false,
	)
	if err != nil {
		if gocerr.GetErrorCode(err) >= http.StatusInternalServerError {
			log.Err(err).
				Ctx(ctx).
				Fields(logFields).
				Msg("[GuestService][countEntities][Count] failed to count entities")
		}

		return 0, err
	}
	logFields["entitiesCount"] = entitiesCount

	if s.cfg.Guest.Cache.Enable && entitiesCount > 0 {
		err = s.setEntitiesCountCache(
			ctx,
			entitiesCountCacheKey,
			entitiesCount,
		)
		if err != nil {
			log.Err(err).
				Ctx(ctx).
				Fields(logFields).
				Msg("[GuestService][countEntities][setEntitiesCountCache] failed to set count entities cache")
			err = nil
		}
	}

	return entitiesCount, nil
}

func (s *GuestService) FindAll(ctx context.Context, requestDTO *dtos.FindAllGuestRequestDTO) (*dtos.FindAllGuestResponseDTO, error) {
	var (
		span                  trace.Span
		logFields             map[string]interface{}
		filter                *goqube.Filter
		sorts                 []goqube.Sort
		listEntityCacheKey    string
		entitiesCountCacheKey string
		errTask               gotask.ErrorTask
		errTaskCtx            context.Context
		listEntity            []entities.GuestEntity
		entitiesCount         uint64
		logLevel              zerolog.Level
		responseDTO           *dtos.FindAllGuestResponseDTO
		err                   error
	)

	ctx, span = tracer.Start(ctx, "[GuestService][FindAll]")
	defer span.End()

	if requestDTO == nil {
		requestDTO = dtos.NewFindAllGuestRequestDTO()
	}

	logFields = map[string]interface{}{
		"requestDTO": requestDTO,
	}

	filter, sorts, err = requestDTO.ToFilterAndSorts()
	if err != nil {
		log.Warn().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestService][FindAll][ToFilter] failed to transform requestDTO into filter and sorts")
		return nil, err
	}
	logFields["filter"] = filter
	logFields["sorts"] = sorts

	listEntityCacheKey = fmt.Sprintf(
		"keyword=%s&sorts=%s&take=%d&skip=%d",
		requestDTO.Keyword,
		requestDTO.Sorts,
		requestDTO.Take,
		requestDTO.Skip,
	)
	listEntityCacheKey = regexp.MustCompile(`[^a-zA-Z0-9:_&=-]+`).
		ReplaceAllString(strings.TrimSpace(listEntityCacheKey), "_")
	listEntityCacheKey = fmt.Sprintf(s.cfg.Guest.Cache.Keyf, listEntityCacheKey)
	logFields["listEntityCacheKey"] = listEntityCacheKey

	entitiesCountCacheKey = fmt.Sprintf("%s:count", listEntityCacheKey)
	logFields["entitiesCountCacheKey"] = entitiesCountCacheKey

	errTask, errTaskCtx = gotask.NewErrorTask(ctx, 2)

	errTask.Go(func() error {
		var errRoutine error

		listEntity, errRoutine = s.findListEntity(
			errTaskCtx,
			listEntityCacheKey,
			filter,
			sorts,
			requestDTO.Take,
			requestDTO.Skip,
		)
		if errRoutine != nil {
			log.Err(errRoutine).
				Ctx(errTaskCtx).
				Fields(logFields).
				Msg("[GuestService][FindAll][findListEntity] failed to find list entity")
			return errRoutine
		}

		return nil
	})

	errTask.Go(func() error {
		var errRoutine error

		entitiesCount, errRoutine = s.countEntities(
			errTaskCtx,
			entitiesCountCacheKey,
			filter,
		)
		if errRoutine != nil {
			logLevel = zerolog.WarnLevel
			if gocerr.GetErrorCode(errRoutine) >= http.StatusInternalServerError {
				logLevel = zerolog.ErrorLevel
			}

			log.WithLevel(logLevel).
				Ctx(errTaskCtx).
				Err(errRoutine).
				Fields(logFields).
				Msg("[GuestService][FindAll][countEntities] failed to count entities")
			return errRoutine
		}

		return nil
	})

	err = errTask.Wait()
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[GuestService][FindAll][Wait] failed to find or count entities")
		return nil, err
	}

	responseDTO = dtos.NewFindAllGuestResponseDTO(listEntity, entitiesCount)

	return responseDTO, nil
}

func (s *GuestService) getEntityByIDCache(ctx context.Context, cacheKey string) (*entities.GuestEntity, error) {
	var (
		span      trace.Span
		logFields map[string]interface{}
		entity    *entities.GuestEntity
		err       error
	)

	ctx, span = tracer.Start(ctx, "[GuestService][getEntityByIDCache]")
	defer span.End()

	logFields = map[string]interface{}{
		"cacheKey": cacheKey,
	}

	entity, err = s.guestCacheRepository.Get(ctx, cacheKey)
	if err != nil {
		if gocerr.GetErrorCode(err) >= http.StatusInternalServerError {
			log.Err(err).
				Ctx(ctx).
				Fields(logFields).
				Msg("[GuestService][getEntityByIDCache][Get] failed to get entity by id cache")
		}

		return nil, err
	}

	return entity, nil
}

func (s *GuestService) setEntityByIDCache(
	ctx context.Context,
	cacheKey string,
	entity *entities.GuestEntity,
) error {
	var (
		span      trace.Span
		logFields map[string]interface{}
		err       error
	)

	ctx, span = tracer.Start(ctx, "[GuestService][setEntityByIDCache]")
	defer span.End()

	logFields = map[string]interface{}{
		"cacheKey":   cacheKey,
		"entity":     entity,
		"expiration": s.cfg.Guest.Cache.Duration,
	}

	err = s.guestCacheRepository.Set(ctx, cacheKey, entity, s.cfg.Guest.Cache.Duration)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[GuestService][setEntityByIDCache][Set] failed to set entity cache")
		return err
	}

	return nil
}

func (s *GuestService) findEntityByID(
	ctx context.Context,
	cacheKey string,
	filter *goqube.Filter,
) (*entities.GuestEntity, error) {
	var (
		span      trace.Span
		logFields map[string]interface{}
		entity    *entities.GuestEntity
		err       error
	)

	ctx, span = tracer.Start(ctx, "[GuestService][findEntityByID]")
	defer span.End()

	logFields = map[string]interface{}{
		"cacheKey": cacheKey,
		"filter":   filter,
	}

	if s.cfg.Guest.Cache.Enable {
		entity, err = s.getEntityByIDCache(ctx, cacheKey)
		if err != nil {
			if gocerr.GetErrorCode(err) >= http.StatusInternalServerError {
				log.Err(err).
					Ctx(ctx).
					Fields(logFields).
					Msg("[GuestService][findEntityByID][getEntityByIDCache] failed to get entity by id cache")
			}

			err = nil
		}

		if entity != nil {
			return entity, nil
		}
	}

	entity, err = s.guestRepository.FindOne(
		ctx,
		filter,
		nil,
		false,
	)
	if err != nil {
		if gocerr.GetErrorCode(err) >= http.StatusInternalServerError {
			log.Err(err).
				Ctx(ctx).
				Fields(logFields).
				Msg("[GuestService][findEntityByID][FindOne] failed to find entity")
		}

		return nil, err
	}
	logFields["entity"] = entity

	if s.cfg.Guest.Cache.Enable {
		err = s.setEntityByIDCache(
			ctx,
			cacheKey,
			entity,
		)
		if err != nil {
			log.Err(err).
				Ctx(ctx).
				Fields(logFields).
				Msg("[GuestService][findEntityByID][setEntityByIDCache] failed to set entity by id cache")
			err = nil
		}
	}

	return entity, nil
}

func (s *GuestService) FindByID(ctx context.Context, requestDTO *dtos.FindGuestByIDRequestDTO) (*dtos.GuestResponseDTO, error) {
	var (
		span        trace.Span
		logFields   map[string]interface{}
		cacheKey    string
		filter      *goqube.Filter
		entity      *entities.GuestEntity
		responseDTO *dtos.GuestResponseDTO
		err         error
	)

	ctx, span = tracer.Start(ctx, "[GuestService][FindByID]")
	defer span.End()

	if requestDTO == nil {
		return nil, gocerr.New(http.StatusBadRequest, "requestDTO is nil")
	}

	logFields = map[string]interface{}{
		"requestDTO": requestDTO,
	}

	err = requestDTO.Validate()
	if err != nil {
		log.Warn().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestService][FindByID][Validate] failed to validate dto")
		return nil, err
	}

	cacheKey = fmt.Sprintf(s.cfg.Guest.Cache.Keyf, requestDTO.ID)
	logFields["cacheKey"] = cacheKey

	filter = s.buildActiveEntityFilterByIDs(requestDTO.ID)
	logFields["filter"] = filter

	entity, err = s.findEntityByID(ctx, cacheKey, filter)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[GuestService][FindByID][FindOne] failed to find entity by id")
		return nil, err
	}

	responseDTO = dtos.NewGuestResponseDTO(entity)

	return responseDTO, nil
}

func (s *GuestService) UpdateByID(ctx context.Context, requestDTO *dtos.UpdateGuestByIDRequestDTO) (*dtos.GuestResponseDTO, error) {
	var (
		span        trace.Span
		logFields   map[string]interface{}
		filter      *goqube.Filter
		entity      *entities.GuestEntity
		logLevel    zerolog.Level
		responseDTO *dtos.GuestResponseDTO
		err         error
	)

	ctx, span = tracer.Start(ctx, "[GuestService][UpdateByID]")
	defer span.End()

	if requestDTO == nil {
		return nil, gocerr.New(http.StatusBadRequest, "requestDTO is nil")
	}

	logFields = map[string]interface{}{
		"requestDTO": requestDTO,
	}

	err = requestDTO.Validate()
	if err != nil {
		log.Warn().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestService][UpdateByID][Validate] failed to validate dto")
		return nil, err
	}

	filter = s.buildActiveEntityFilterByIDs(requestDTO.ID)
	logFields["filter"] = filter

	entity, err = s.guestRepository.FindOne(ctx, filter, nil, false)
	if err != nil {
		logLevel = zerolog.WarnLevel
		if gocerr.GetErrorCode(err) >= http.StatusInternalServerError {
			logLevel = zerolog.ErrorLevel
		}

		log.WithLevel(logLevel).
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestService][UpdateByID][FindOne] failed to find entity")
		return nil, err
	}

	entity = requestDTO.ToExistingEntity(entity)
	logFields["entity"] = entity

	err = s.withTransaction(ctx, logFields, "UpdateByID", func(tx repositories.IBoilerplateDatabaseTransaction) error {
		return s.guestRepository.WithTransaction(tx).Update(ctx, entity, filter)
	})
	if err != nil {
		return nil, err
	}

	responseDTO = dtos.NewGuestResponseDTO(entity)
	logFields["responseDTO"] = responseDTO

	s.tryDeleteEntityCaches(ctx, logFields, "UpdateByID")
	s.publishEvent(ctx, logFields, s.cfg.Guest.Event.Updated.Enable, s.cfg.Guest.Event.Updated.Topic, "UpdateByID", *entity)

	return responseDTO, nil
}

func (s *GuestService) ProcessEvent(ctx context.Context, requestDTO *dtos.GuestEventRequestDTO) (*dtos.GuestEventResponseDTO, error) {
	var (
		span        trace.Span
		logFields   map[string]interface{}
		entity      *entities.GuestEventEntity
		responseDTO *dtos.GuestEventResponseDTO
		err         error
	)

	ctx, span = tracer.Start(ctx, "[GuestService][ProcessEvent]")
	defer span.End()

	if requestDTO == nil {
		return nil, gocerr.New(http.StatusBadRequest, "requestDTO is nil")
	}

	logFields = map[string]interface{}{
		"requestDTO": requestDTO,
	}

	entity = requestDTO.ToEntity()
	logFields["entity"] = entity

	err = s.webhookSiteRepository.SendWebhook(ctx, entity)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[GuestService][ProcessEvent][SendWebhook] failed to send webhook")
		return nil, err
	}

	responseDTO = dtos.NewGuestEventResponseDTO(entity)

	return responseDTO, nil
}

func (s *GuestService) BulkCreate(ctx context.Context, requestDTO *dtos.BulkCreateGuestsRequestDTO) (*dtos.BulkCreateGuestsResponseDTO, error) {
	var (
		span        trace.Span
		logFields   map[string]interface{}
		newEntities []entities.GuestEntity
		responseDTO *dtos.BulkCreateGuestsResponseDTO
		err         error
	)

	ctx, span = tracer.Start(ctx, "[GuestService][BulkCreate]")
	defer span.End()

	if requestDTO == nil {
		return nil, gocerr.New(http.StatusBadRequest, "requestDTO is nil")
	}

	logFields = map[string]interface{}{
		"requestDTO": requestDTO,
	}

	err = requestDTO.Validate()
	if err != nil {
		log.Warn().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestService][BulkCreate][Validate] failed to validate dto")
		return nil, err
	}

	newEntities = requestDTO.ToEntities()
	logFields["newEntities"] = newEntities

	err = s.withTransaction(ctx, logFields, "BulkCreate", func(tx repositories.IBoilerplateDatabaseTransaction) error {
		return s.guestRepository.WithTransaction(tx).BulkCreate(ctx, newEntities)
	})
	if err != nil {
		return nil, err
	}

	responseDTO = dtos.NewBulkCreateGuestsResponseDTO(newEntities)
	logFields["responseDTO"] = responseDTO

	s.tryDeleteEntityCaches(ctx, logFields, "BulkCreate")
	s.publishEvent(ctx, logFields, s.cfg.Guest.Event.BulkCreated.Enable, s.cfg.Guest.Event.BulkCreated.Topic, "BulkCreate", newEntities...)

	return responseDTO, nil
}

func (s *GuestService) BulkUpdate(ctx context.Context, requestDTO *dtos.BulkUpdateGuestsRequestDTO) (*dtos.BulkUpdateGuestsResponseDTO, error) {
	var (
		span                trace.Span
		logFields           map[string]interface{}
		entityIDs           []string
		filter              *goqube.Filter
		existingEntities    []entities.GuestEntity
		existingEntitiesMap map[string]*entities.GuestEntity
		existingEntity      *entities.GuestEntity
		ok                  bool
		updatedEntity       *entities.GuestEntity
		updatedEntities     []entities.GuestEntity
		responseDTO         *dtos.BulkUpdateGuestsResponseDTO
		err                 error
	)

	ctx, span = tracer.Start(ctx, "[GuestService][BulkUpdate]")
	defer span.End()

	if requestDTO == nil {
		return nil, gocerr.New(http.StatusBadRequest, "requestDTO is nil")
	}

	logFields = map[string]interface{}{
		"requestDTO": requestDTO,
	}

	err = requestDTO.Validate()
	if err != nil {
		log.Warn().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestService][BulkUpdate][Validate] failed to validate dto")
		return nil, err
	}

	entityIDs = requestDTO.ToIDs()

	filter = s.buildActiveEntityFilterByIDs(entityIDs...)
	logFields["filter"] = filter

	existingEntities, err = s.guestRepository.FindAll(ctx, filter, nil, uint64(len(requestDTO.Items)), 0, false)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[GuestService][BulkUpdate][FindAll] failed to find entities")
		return nil, err
	}

	if len(existingEntities) <= 0 {
		return nil, gocerr.New(http.StatusNotFound, "entities not found")
	}
	logFields["existingEntities"] = existingEntities

	existingEntitiesMap = map[string]*entities.GuestEntity{}
	for i := range existingEntities {
		existingEntitiesMap[existingEntities[i].ID.String()] = &existingEntities[i]
	}
	logFields["existingEntitiesMap"] = existingEntitiesMap

	for i := range requestDTO.Items {
		var item = &requestDTO.Items[i]

		existingEntity, ok = existingEntitiesMap[item.ID]
		if !ok {
			err = gocerr.New(http.StatusNotFound, "entity not found for id: "+item.ID)
			return nil, err
		}

		updatedEntity = item.ToExistingEntity(existingEntity)
		updatedEntities = append(updatedEntities, *updatedEntity)
	}
	logFields["updatedEntities"] = updatedEntities

	err = s.withTransaction(ctx, logFields, "BulkUpdate", func(tx repositories.IBoilerplateDatabaseTransaction) error {
		return s.guestRepository.WithTransaction(tx).BulkUpdate(ctx, updatedEntities)
	})
	if err != nil {
		return nil, err
	}

	responseDTO = dtos.NewBulkUpdateGuestsResponseDTO(updatedEntities)
	logFields["responseDTO"] = responseDTO

	s.tryDeleteEntityCaches(ctx, logFields, "BulkUpdate")
	s.publishEvent(ctx, logFields, s.cfg.Guest.Event.BulkUpdated.Enable, s.cfg.Guest.Event.BulkUpdated.Topic, "BulkUpdate", updatedEntities...)

	return responseDTO, nil
}

func (s *GuestService) BulkDelete(ctx context.Context, requestDTO *dtos.BulkDeleteGuestsRequestDTO) error {
	var (
		span             trace.Span
		logFields        map[string]interface{}
		entityIDs        []string
		filter           *goqube.Filter
		existingEntities []entities.GuestEntity
		deletedEntities  []entities.GuestEntity
		err              error
	)

	ctx, span = tracer.Start(ctx, "[GuestService][BulkDelete]")
	defer span.End()

	if requestDTO == nil {
		return gocerr.New(http.StatusBadRequest, "requestDTO is nil")
	}

	logFields = map[string]interface{}{
		"requestDTO": requestDTO,
	}

	err = requestDTO.Validate()
	if err != nil {
		log.Warn().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[GuestService][BulkDelete][Validate] failed to validate dto")
		return err
	}

	entityIDs = requestDTO.ToIDs()

	filter = s.buildActiveEntityFilterByIDs(entityIDs...)
	logFields["filter"] = filter

	existingEntities, err = s.guestRepository.FindAll(ctx, filter, nil, uint64(len(requestDTO.IDs)), 0, false)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[GuestService][BulkDelete][FindAll] failed to find entities")
		return err
	}

	if len(existingEntities) <= 0 {
		return gocerr.New(http.StatusNotFound, "entities not found")
	}
	logFields["existingEntities"] = existingEntities

	for i := range existingEntities {
		deletedEntities = append(deletedEntities, *existingEntities[i].MarkAsDeleted(requestDTO.DeletedBy))
	}
	logFields["deletedEntities"] = deletedEntities

	err = s.withTransaction(ctx, logFields, "BulkDelete", func(tx repositories.IBoilerplateDatabaseTransaction) error {
		return s.guestRepository.WithTransaction(tx).BulkUpdate(ctx, deletedEntities)
	})
	if err != nil {
		return err
	}

	s.tryDeleteEntityCaches(ctx, logFields, "BulkDelete")
	s.publishEvent(ctx, logFields, s.cfg.Guest.Event.BulkDeleted.Enable, s.cfg.Guest.Event.BulkDeleted.Topic, "BulkDelete", deletedEntities...)

	return nil
}
