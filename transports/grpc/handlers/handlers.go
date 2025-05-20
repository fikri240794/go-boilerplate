package handlers

import (
	"go-boilerplate/internal/services"
	"go-boilerplate/pkg/protobuf_boilerplate"
)

type ImplementedBoilerplateServer struct {
	protobuf_boilerplate.UnimplementedBoilerplateServer

	guestService services.IGuestService
}

func NewImplementedBoilerplateServer(
	guestService services.IGuestService,
) *ImplementedBoilerplateServer {
	return &ImplementedBoilerplateServer{
		guestService: guestService,
	}
}
