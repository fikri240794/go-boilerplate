package handlers

import "github.com/gofiber/fiber/v2"

type Handlers struct {
	Guest *GuestHandler
}

func (r *Handlers) SetupRoutes(server *fiber.App) {
	r.Guest.SetupRoutes(server)
}
