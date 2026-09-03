package auth

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mohammed-ayoub-dz/hifz/handlers"
)

func RegisterRoutes(router fiber.Router) {
	authGroup := router.Group("/auth")

	authGroup.Post("/google", handlers.GoogleLogin)
}