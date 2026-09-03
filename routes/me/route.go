package me

import (
	"github.com/gofiber/fiber/v3"

	"github.com/mohammed-ayoub-dz/hifz/config"
	"github.com/mohammed-ayoub-dz/hifz/handlers"
	"github.com/mohammed-ayoub-dz/hifz/middleware"
)

func RegisterRoutes(api fiber.Router) {
	me := api.Group(
		"/me",
		middleware.Protected(),
	)

	me.Get("/", handlers.GetMe(config.DB))
}