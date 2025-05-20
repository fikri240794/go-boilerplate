package handlers

import (
	"context"
	"go-boilerplate/pkg/constants"
	"go-boilerplate/transports/event_consumer/models/vms"
	"net/http"

	"github.com/fikri240794/gocerr"
	"github.com/goccy/go-json"
	"github.com/nsqio/go-nsq"
	"github.com/rs/zerolog/log"
)

type messageHandler struct {
	handleMessageFunc func(ctx context.Context, m *nsq.Message) error
}

func NewMessageHandler(handleMessageFunc func(ctx context.Context, m *nsq.Message) error) nsq.Handler {
	return &messageHandler{
		handleMessageFunc: handleMessageFunc,
	}
}

func (h *messageHandler) HandleMessage(m *nsq.Message) error {
	var (
		ctx       context.Context
		logFields map[string]interface{}
		requestVM *vms.EventRequestVM[interface{}]
		err       error
	)

	ctx = context.TODO()

	logFields = map[string]interface{}{
		"messageBody": string(m.Body),
	}

	requestVM = &vms.EventRequestVM[interface{}]{}
	err = json.Unmarshal(m.Body, requestVM)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", string(m.ID[:])).
			Msg("[messageHandler][HandleMessage][Unmarshal] failed to parse message body")
		log.Debug().
			Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[messageHandler][HandleMessage][Unmarshal] failed to parse message body")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx = requestVM.ExtractTracerPropagator(ctx)

	ctx = context.WithValue(ctx, constants.ContextKeyRequestID, string(m.ID[:]))

	return h.handleMessageFunc(ctx, m)
}
