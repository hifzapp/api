package handlers

import (

	"github.com/gofiber/fiber/v3"
	"github.com/mohammed-ayoub-dz/hifz/models"
	"gorm.io/gorm"
)

type OnboardingRequest struct {
	DailyGoal int `json:"daily_goal"`
}

const (
	minDailyGoal = 1
	maxDailyGoal = 1000
)

func CompleteOnboarding(db *gorm.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		userIDValue := c.Locals("user_id")

		if userIDValue == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required.",
			})
		}

		userID, ok := userIDValue.(uint)

		if !ok || userID == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authentication context.",
			})
		}

		var req OnboardingRequest

		if err := c.Bind().Body(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body.",
			})
		}

		if req.DailyGoal < minDailyGoal ||
			req.DailyGoal > maxDailyGoal {

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Daily goal must be between 1 and 1000.",
			})
		}

		result := db.
			Model(&models.User{}).
			Where("id = ?", userID).
			Updates(map[string]interface{}{
				"daily_goal": req.DailyGoal,
				"onboarded":  true,
			})

		if result.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save onboarding data.",
			})
		}

		if result.RowsAffected == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found.",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":    "Onboarding completed successfully.",
			"daily_goal": req.DailyGoal,
			"onboarded":  true,
		})
	}
}
