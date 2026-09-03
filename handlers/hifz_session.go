package handlers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/mohammed-ayoub-dz/hifz/config"
	"github.com/mohammed-ayoub-dz/hifz/models"
	"github.com/mohammed-ayoub-dz/hifz/utils"
	"gorm.io/gorm"
)

const (
	defaultSessionPage  = 1
	defaultSessionLimit  = 10
	maxSessionLimit      = 100
	maxSessionDuration   = 24 * 60 * 60
	maxSessionScore      = 100
	maxSessionMistakes   = 10000
	maxSessionHeartsLost = 10000
)

type CreateSessionInput struct {
    SessionType string `json:"session_type"`
    SurahNumber int    `json:"surah_number"`
    StartAyah   int    `json:"start_ayah"`
    EndAyah     int    `json:"end_ayah"`
}

type UpdateSessionInput struct {
	SessionType string `json:"session_type"`

	Duration int `json:"duration"`

	Score      int `json:"score"`
	Mistakes   int `json:"mistakes"`
	HeartsLost int `json:"hearts_lost"`
	IsComplete   bool `json:"is_complete"`

}


type CompleteHifzSessionInput struct {
	Mistakes   int `json:"mistakes"`
	HeartsLost int `json:"hearts_lost"`
}

func CreateHifzSession(c fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required.",
		})
	}

	var input CreateSessionInput

	if err := c.Bind().JSON(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body.",
		})
	}

	normalizeSessionInput(&input)

	if err := validateCreateSessionInput(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	session := models.HifzSession{
    UserID:      userID,
    SessionType: input.SessionType,
    SurahNumber: input.SurahNumber,
    StartAyah:   input.StartAyah,
    EndAyah:     input.EndAyah,
	IsComplete : false,
    }

	if err := config.DB.Create(&session).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create the session.",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Session created successfully.",
		"session": session,
	})
}

func GetUserHifzSessions(c fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required.",
		})
	}

	page, limit, err := parsePagination(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	offset := (page - 1) * limit

	var sessions []models.HifzSession
	var total int64

	query := config.DB.
		Model(&models.HifzSession{}).
		Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count sessions.",
		})
	}

	if err := query.
		Order("created_at DESC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&sessions).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch sessions.",
		})
	}

	totalPages := int64(0)

	if total > 0 {
		totalPages = (total + int64(limit) - 1) / int64(limit)
	}

	return c.JSON(fiber.Map{
		"sessions": sessions,
		"meta": fiber.Map{
			"total":        total,
			"page":         page,
			"limit":        limit,
			"total_pages":  totalPages,
		},
	})
}

func GetHifzSessionByID(c fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required.",
		})
	}

	sessionID, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid session ID.",
		})
	}

	var session models.HifzSession

	err = config.DB.
		Where("id = ? AND user_id = ?", sessionID, userID).
		First(&session).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Session not found.",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch the session.",
		})
	}

	return c.JSON(fiber.Map{
		"session": session,
	})
}

func UpdateHifzSession(c fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required.",
		})
	}

	sessionID, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid session ID.",
		})
	}

	var session models.HifzSession

	err = config.DB.
		Where("id = ? AND user_id = ?", sessionID, userID).
		First(&session).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Session not found.",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch the session.",
		})
	}

	var input UpdateSessionInput

	if err := c.Bind().JSON(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body.",
		})
	}

	normalizeUpdateSessionInput(&input)

	if err := validateUpdateSessionInput(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	updates := map[string]interface{}{
		"session_type": input.SessionType,
		"duration":     input.Duration,
		"score":        input.Score,
		"mistakes":     input.Mistakes,
		"hearts_lost":  input.HeartsLost,
	}

	if err := config.DB.
		Model(&session).
		Updates(updates).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update the session.",
		})
	}

	if err := config.DB.
		Where("id = ? AND user_id = ?", sessionID, userID).
		First(&session).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve the updated session.",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Session updated successfully.",
		"session": session,
	})
}

func DeleteHifzSession(c fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required.",
		})
	}

	sessionID, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid session ID.",
		})
	}

	result := config.DB.
		Where("id = ? AND user_id = ?", sessionID, userID).
		Delete(&models.HifzSession{})

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete the session.",
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Session not found.",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Session deleted successfully.",
	})
}


func CompleteHifzSession(c fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required.",
		})
	}

	sessionID, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid session ID.",
		})
	}

	var input CompleteHifzSessionInput

	if err := c.Bind().JSON(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body.",
		})
	}

	if input.Mistakes < 0 || input.HeartsLost < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid mistakes or hearts lost value.",
		})
	}

	err = config.DB.Transaction(func(tx *gorm.DB) error {
		var session models.HifzSession

		err := tx.
			Where(
				"id = ? AND user_id = ? AND is_complete = ?",
				sessionID,
				userID,
				false,
			).
			First(&session).Error

		if err != nil {
			return err
		}

		result := tx.
			Model(&session).
			Updates(map[string]interface{}{
				"is_complete": true,
				"mistakes":    input.Mistakes,
				"hearts_lost": input.HeartsLost,
			})

		if result.Error != nil {
			return result.Error
		}

		result = tx.
			Model(&models.User{}).
			Where("id = ?", userID).
			Updates(map[string]interface{}{
				"hearts": gorm.Expr(
					"hearts - ?",
					input.HeartsLost,
				),
				"streak": gorm.Expr(
					"streak + ?",
					1,
				),
			})

		if result.Error != nil {
			return result.Error
		}

		return nil
	})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Hifz session not found or already completed.",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to complete hifz session.",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Hifz session completed successfully.",
	})
}

func parseID(value string) (uint, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("invalid ID")
	}

	return uint(id), nil
}

func parsePagination(c fiber.Ctx) (int, int, error) {
	page := defaultSessionPage
	limit := defaultSessionLimit

	if value := c.Query("page"); value != "" {
		parsed, err := strconv.Atoi(value)

		if err != nil || parsed < 1 {
			return 0, 0, errors.New("page must be a positive integer")
		}

		page = parsed
	}

	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)

		if err != nil || parsed < 1 {
			return 0, 0, errors.New("limit must be a positive integer")
		}

		if parsed > maxSessionLimit {
			return 0, 0, errors.New("limit cannot exceed 100")
		}

		limit = parsed
	}

	return page, limit, nil
}

func normalizeSessionInput(input *CreateSessionInput) {
	input.SessionType = strings.ToLower(strings.TrimSpace(input.SessionType))
}

func normalizeUpdateSessionInput(input *UpdateSessionInput) {
	input.SessionType = strings.ToLower(strings.TrimSpace(input.SessionType))
}

func validateCreateSessionInput(input CreateSessionInput) error {
	switch input.SessionType {
	case "memorization", "review":
	default:
		return errors.New(
			"session_type must be either 'memorization' or 'review'",
		)
	}



	if input.SurahNumber < 1 || input.SurahNumber > 114 {
		return errors.New("surah_number must be between 1 and 114")
	}

	if input.StartAyah < 1 {
		return errors.New("start_ayah must be greater than zero")
	}

	if input.EndAyah < input.StartAyah {
		return errors.New("end_ayah must be greater than or equal to start_ayah")
	}


	return nil
}

func validateUpdateSessionInput(input UpdateSessionInput) error {
	switch input.SessionType {
	case "memorization", "review":
	default:
		return errors.New(
			"session_type must be either 'memorization' or 'review'",
		)
	}

	if input.Duration <= 0 {
		return errors.New("duration must be greater than zero")
	}

	if input.Duration > maxSessionDuration {
		return errors.New("duration cannot exceed 24 hours")
	}

	if input.Score < 0 || input.Score > maxSessionScore {
		return errors.New("score must be between 0 and 100")
	}

	if input.Mistakes < 0 || input.Mistakes > maxSessionMistakes {
		return errors.New("mistakes must be between 0 and 10000")
	}

	if input.HeartsLost < 0 || input.HeartsLost > maxSessionHeartsLost {
		return errors.New("hearts_lost must be between 0 and 10000")
	}

	return nil
}
