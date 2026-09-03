package dashboard

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mohammed-ayoub-dz/hifz/handlers"
)

func RegisterRoutes(router fiber.Router) {
	router.Get("/dashboard", handlers.GetDashboard)
}