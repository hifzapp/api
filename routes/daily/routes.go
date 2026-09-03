package daily

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mohammed-ayoub-dz/hifz/handlers"
)

func RegisterRoutes(router fiber.Router) {
	progress := router.Group("/daily-progress")
	progress.Get("/", handlers.GetUserDailyProgress)
	progress.Post("/", handlers.CreateDailyProgress)
	progress.Get("/:date", handlers.GetDailyProgressByDate)
	progress.Patch("/:date", handlers.UpdateDailyProgress)
	progress.Delete("/:date", handlers.DeleteDailyProgress)
}
