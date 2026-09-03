package utils

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/mohammed-ayoub-dz/hifz/middleware"
)

func GetUserID(c fiber.Ctx) (uint, error) {
	val := c.Locals(middleware.UserIDKey)
	userID, ok := val.(uint)
	if !ok || userID == 0 {
		return 0, errors.New("unauthorized: user id not found in context")
	}
	return userID, nil
}