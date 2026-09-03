package middleware

import (
	"errors"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

const UserIDKey = "user_id"

type AuthClaims struct {
	UserID uint `json:"user_id"`
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

		c.Locals(UserIDKey, claims.UserID)

		return c.Next()
	}
}