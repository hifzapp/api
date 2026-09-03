package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mohammed-ayoub-dz/hifz/config"
	"github.com/mohammed-ayoub-dz/hifz/models"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

type GoogleAuthRequest struct {
	Token string `json:"token"`
}

type AuthClaims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

type AuthUserResponse struct {
	ID     uint   `json:"id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Hearts int    `json:"hearts"`
	Streak int    `json:"streak"`
}

func GoogleLogin(c fiber.Ctx) error {

	var req GoogleAuthRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body.",
		})
	}

	req.Token = strings.TrimSpace(req.Token)

	if req.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "A Google token is required.",
		})
	}


	googleClientID := strings.TrimSpace(
		os.Getenv("GOOGLE_CLIENT_ID"),
	)

	jwtSecret := os.Getenv("JWT_SECRET")

	if googleClientID == "" || len(jwtSecret) < 32 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Authentication service is not configured.",
		})
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	payload, err := idtoken.Validate(
		ctx,
		req.Token,
		googleClientID,
	)

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Google token is invalid or expired.",
		})
	}


	googleID := strings.TrimSpace(payload.Subject)

	if googleID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid Google identity.",
		})
	}

	emailClaim, ok := payload.Claims["email"]

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Google account email is missing.",
		})
	}

	email, ok := emailClaim.(string)

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid Google account email.",
		})
	}

	email = strings.ToLower(strings.TrimSpace(email))

	if email == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Google account email is missing.",
		})
	}

	// 
	
	emailVerifiedClaim, ok := payload.Claims["email_verified"]

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Google email verification status is missing.",
		})
	}

	emailVerified, ok := emailVerifiedClaim.(bool)

	if !ok || !emailVerified {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Google email is not verified.",
		})
	}


	name := ""

	if value, ok := payload.Claims["name"].(string); ok {
		name = strings.TrimSpace(value)
	}

	avatar := ""

	if value, ok := payload.Claims["picture"].(string); ok {
		avatar = strings.TrimSpace(value)
	}


	var user models.User

	err = config.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.
			Where("google_id = ?", googleID).
			First(&user)

		if result.Error == nil {
			return tx.
				Model(&user).
				Updates(map[string]interface{}{
					"name":   name,
					"avatar": avatar,
				}).Error
		}

		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return result.Error
		}

		result = tx.
			Where("email = ?", email).
			First(&user)

		if result.Error == nil {

			return tx.
				Model(&user).
				Updates(map[string]interface{}{
					"google_id": googleID,
					"name":      name,
					"avatar":    avatar,
				}).Error
		}

		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return result.Error
		}

		user = models.User{
			GoogleID: googleID,
			Email:    email,
			Name:     name,
			Avatar:   avatar,
			Hearts:   5,
			Streak:   0,
		}

		return tx.Create(&user).Error
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to complete authentication.",
		})
	}


	now := time.Now()

	tokenID, err := generateTokenID()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to create authentication session.",
		})
	}

	claims := AuthClaims{
		UserID: user.ID,

		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "hifz-api",

			Subject: googleID,

			Audience: jwt.ClaimStrings{
				"hifz-client",
			},

			IssuedAt: jwt.NewNumericDate(now),

			NotBefore: jwt.NewNumericDate(now),

			ExpiresAt: jwt.NewNumericDate(
				now.Add(72 * time.Hour),
			),

			ID: tokenID,
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	tokenString, err := token.SignedString(
		[]byte(jwtSecret),
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to create authentication session.",
		})
	}


	isProd := isProduction();

	sameSite := "Lax"

	if isProd {
		sameSite = "None"
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  now.Add(72 * time.Hour),
		HTTPOnly: true,
		Secure:   isProd,
		SameSite: sameSite,
		Path:     "/",
	})


	userResponse := AuthUserResponse{
		ID:     user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Avatar: user.Avatar,
		Hearts: user.Hearts,
		Streak: user.Streak,
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged in successfully.",
		"user":    userResponse,
	})
}

func isProduction() bool {
    return os.Getenv("APP_ENV") == "production"
}


func generateTokenID() (string, error) {
	var b [32]byte

	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}

	return hex.EncodeToString(b[:]), nil
}

