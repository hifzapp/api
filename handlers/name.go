package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/mohammed-ayoub-dz/hifz/models"
	"gorm.io/gorm"
	"unicode/utf8"
)

type UpdateNameRequest struct {
	Name string `json:"name"`
}

const (
	minNameLength = 1
	maxNameLength = 14
)

func UpdateName(db *gorm.DB) fiber.Handler {
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

		var req UpdateNameRequest

		if err := c.Bind().Body(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body.",
			})
		}

		req.Name = strings.TrimSpace(req.Name)

		if err := validateName(req.Name); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		result := db.
			Model(&models.User{}).
			Where("id = ?", userID).
			Update("name", req.Name)

		if result.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update name.",
			})
		}

		if result.RowsAffected == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found.",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Name updated successfully.",
			"name":    req.Name,
		})
	}
}

func validateName(name string) error {
    length := utf8.RuneCountInString(name)

    if length < minNameLength {
        return errors.New("Name is required.")
    }

    if length > maxNameLength {
        return errors.New("Name cannot exceed 14 characters.")
    }

    return nil
}