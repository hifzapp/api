package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/mohammed-ayoub-dz/hifz/models"
	"gorm.io/gorm"
)

func GetMe(db *gorm.DB) fiber.Handler {
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
		var user models.User

		err := db.
			Select(
				"id",
				"name",
				"avatar",
				"daily_goal",
				"onboarded",
				"hearts",
				"streak",
				"created_at",
			).
			Where("id = ?", userID).
			First(&user).
			Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "User not found.",
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve user profile.",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"user": user,
		})
	}
}