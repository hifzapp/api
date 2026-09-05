package middleware

import (
	"errors"
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mohammed-ayoub-dz/hifz/config"
	"github.com/mohammed-ayoub-dz/hifz/models"
)

const UserIDKey = "user_id"

type AuthClaims struct {
	UserID         uint `json:"user_id"`
	SessionVersion int  `json:"session_version"`
	jwt.RegisteredClaims
}

func Protected() fiber.Handler {
	secret := os.Getenv("JWT_SECRET")

	return func(c fiber.Ctx) error {
		if secret == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Authentication service is not configured.",
			})
		}

		tokenString := c.Cookies("token")

		if tokenString == "" {
			if authorization := c.Get("Authorization"); authorization != "" {
				tokenString = strings.TrimPrefix(authorization, "Bearer ")
			}
		}

		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required.",
			})
		}

		token, err := jwt.ParseWithClaims(
			tokenString,
			&AuthClaims{},
			func(token *jwt.Token) (any, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, errors.New("unexpected signing method")
				}

				return []byte(secret), nil
			},
		)

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired authentication session.",
			})
		}

		claims, ok := token.Claims.(*AuthClaims)

		if !ok || claims.UserID == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authentication data.",
			})
		}

		var user models.User
		if err := config.DB.First(&user, claims.UserID).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User session is no longer valid.",
			})
		}

		if user.SessionVersion != claims.SessionVersion {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Your session has been invalidated.",
			})
		}

		c.Locals(UserIDKey, claims.UserID)

		return c.Next()
	}
}
