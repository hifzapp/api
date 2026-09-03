package sessions

import (
	"github.com/gofiber/fiber/v3"

	"github.com/mohammed-ayoub-dz/hifz/handlers"
)

func RegisterRoutes(router fiber.Router) {
	sessions := router.Group("/sessions")

	sessions.Get("/", handlers.GetUserHifzSessions)
	sessions.Post("/", handlers.CreateHifzSession)

	sessions.Get("/:id", handlers.GetHifzSessionByID)
	sessions.Patch("/:id", handlers.UpdateHifzSession)
	sessions.Delete("/:id", handlers.DeleteHifzSession)

	sessions.Post("/:id/complete", handlers.CompleteHifzSession)
}
