package repositories

import (
	"context"
	"fmt"
	"go-boilerplate/configs"
	"go-boilerplate/datasources/webhook_site_http_client"
	"go-boilerplate/internal/models/entities"
	"go-boilerplate/pkg/constants"
	custom_context "go-boilerplate/pkg/context"
	"go-boilerplate/pkg/tracer"
	"net/http"

	"github.com/fikri240794/gocerr"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type IWebhookSiteRepository interface {
	SendWebhook(ctx context.Context, requestData *entities.GuestEventEntity) error
}

type WebhookSiteRepository struct {
	cfg        *configs.Config
	httpClient *webhook_site_http_client.WebhookSiteHTTPClient
}

func NewWebhookSiteRepository(cfg *configs.Config, httpClient *webhook_site_http_client.WebhookSiteHTTPClient) *WebhookSiteRepository {
	return &WebhookSiteRepository{
		cfg:        cfg,
		httpClient: httpClient,
	}
}

func (r *WebhookSiteRepository) SendWebhook(ctx context.Context, requestData *entities.GuestEventEntity) error {
	var (
		span               trace.Span
		logFields          map[string]interface{}
		logLevel           zerolog.Level
		httpRequestHeaders map[string]string
		httpResponse       *resty.Response
		err                error
	)

	ctx, span = tracer.Start(ctx, "[WebhookSiteRepository][SendWebhook]")
	defer span.End()

	httpRequestHeaders = map[string]string{
		"Content-Type": "application/json",
	}

	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(httpRequestHeaders))

	logFields = map[string]interface{}{
		"requestid": custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID),
		"url": fmt.Sprintf(
			"%s%s",
			r.cfg.Datasource.WebhookSiteHTTPClient.BaseURL,
			r.cfg.Datasource.WebhookSiteHTTPClient.Endpoint.Webhook,
		),
		"requestHeaders": httpRequestHeaders,
		"requestData":    requestData,
		"requestMethod":  http.MethodPost,
	}

	log.Debug().
		Ctx(ctx).
		Fields(logFields).
		Msg("[WebhookSiteRepository][SendWebhook] http request in progress")

	httpResponse, err = r.httpClient.HttpClient.R().
		SetContext(ctx).
		SetHeaders(httpRequestHeaders).
		SetBody(requestData).
		Post(r.cfg.Datasource.WebhookSiteHTTPClient.Endpoint.Webhook)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[WebhookSiteRepository][SendWebhook][Post] failed to request http")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[WebhookSiteRepository][SendWebhook][Post] failed to request http")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	logFields["statusCode"] = httpResponse.StatusCode()
	logFields["responseHeaders"] = httpResponse.Header()
	logFields["responseBody"] = string(httpResponse.Body())

	log.Debug().
		Ctx(ctx).
		Fields(logFields).
		Msg("[WebhookSiteRepository][SendWebhook] http request in completed")

	if httpResponse.StatusCode() >= http.StatusBadRequest {
		logLevel = zerolog.WarnLevel
		err = gocerr.New(httpResponse.StatusCode(), string(httpResponse.Body()))

		if httpResponse.StatusCode() >= http.StatusInternalServerError {
			logLevel = zerolog.ErrorLevel
		}
	}

	if err != nil {
		log.WithLevel(logLevel).
			Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msgf("[WebhookSiteRepository][SendWebhook] http response is not success")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msgf("[WebhookSiteRepository][SendWebhook] http response is not success")
		return err
	}

	return nil
}
