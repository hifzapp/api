package hifz

import (
	"github.com/gofiber/fiber/v3"

	"github.com/mohammed-ayoub-dz/hifz/handlers"
)

func RegisterRoutes(router fiber.Router) {
	progress := router.Group("/hifz/progress")

	progress.Get("/", handlers.GetUserHifzProgress)
	progress.Post("/", handlers.CreateHifzProgress)

	progress.Get("/:id", handlers.GetHifzProgressByID)
	progress.Patch("/:id", handlers.UpdateHifzProgress)
	progress.Delete("/:id", handlers.DeleteHifzProgress)
}
