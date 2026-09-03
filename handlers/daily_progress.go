package handlers

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/mohammed-ayoub-dz/hifz/config"
	"github.com/mohammed-ayoub-dz/hifz/models"
	"github.com/mohammed-ayoub-dz/hifz/utils"
	"gorm.io/gorm"
)

const (
	dateLayout          = "2006-01-02"
	maxDailyAyahs       = 1000
	maxDailyGoalTarget  = 1000
)

type CreateDailyProgressInput struct {
	Date           string `json:"date"`
	NewAyahsCount  int    `json:"new_ayahs_count"`
	ReviewedCount  int    `json:"reviewed_count"`
	GoalTarget     int    `json:"goal_target"`
	IsGoalAchieved bool   `json:"is_goal_achieved"`
}

type UpdateDailyProgressInput struct {
	NewAyahsCount  int  `json:"new_ayahs_count"`
	ReviewedCount  int  `json:"reviewed_count"`
	GoalTarget     int  `json:"goal_target"`
	IsGoalAchieved bool `json:"is_goal_achieved"`
}

func CreateDailyProgress(c fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required.",
		})
	}

	var input CreateDailyProgressInput

	if err := c.Bind().JSON(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body.",
		})
	}

	input.Date = strings.TrimSpace(input.Date)

	date, err := time.Parse(dateLayout, input.Date)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "date must use the YYYY-MM-DD format.",
		})
	}

	if err := validateDailyProgress(
		input.NewAyahsCount,
		input.ReviewedCount,
		input.GoalTarget,
	); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var existing models.DailyProgress

	err = config.DB.
		Where("user_id = ? AND date = ?", userID, date).
		First(&existing).Error

	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Daily progress for this date already exists.",
		})
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check existing daily progress.",
		})
	}

	progress := models.DailyProgress{
		UserID:         userID,
		Date:           date,
		NewAyahsCount:  input.NewAyahsCount,
		ReviewedCount:  input.ReviewedCount,
		GoalTarget:     input.GoalTarget,
		IsGoalAchieved: input.IsGoalAchieved,
	}

	if err := config.DB.Create(&progress).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create daily progress.",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "Daily progress created successfully.",
		"progress": progress,
	})
}

func GetUserDailyProgress(c fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required.",
		})
	}

	page, limit, err := parseDailyPagination(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	offset := (page - 1) * limit

	var progress []models.DailyProgress
	var total int64

	query := config.DB.
		Model(&models.DailyProgress{}).
		Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count daily progress records.",
		})
	}

	if err := query.
		Order("date DESC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&progress).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch daily progress.",
		})
	}

	totalPages := int64(0)

	if total > 0 {
		totalPages = (total + int64(limit) - 1) / int64(limit)
	}

	return c.JSON(fiber.Map{
		"progress": progress,
		"meta": fiber.Map{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
		},
	})
}

func GetDailyProgressByDate(c fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required.",
		})
	}

	date, err := parseDate(c.Params("date"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid date. Use YYYY-MM-DD format.",
		})
	}

	var progress models.DailyProgress

	err = config.DB.
		Where("user_id = ? AND date = ?", userID, date).
		First(&progress).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Daily progress not found.",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch daily progress.",
		})
	}

	return c.JSON(fiber.Map{
		"progress": progress,
	})
}

func UpdateDailyProgress(c fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required.",
		})
	}

	date, err := parseDate(c.Params("date"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid date. Use YYYY-MM-DD format.",
		})
	}

	var progress models.DailyProgress

	err = config.DB.
		Where("user_id = ? AND date = ?", userID, date).
		First(&progress).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Daily progress not found.",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch daily progress.",
		})
	}

	var input UpdateDailyProgressInput

	if err := c.Bind().JSON(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body.",
		})
	}

	if err := validateDailyProgress(
		input.NewAyahsCount,
		input.ReviewedCount,
		input.GoalTarget,
	); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	updates := map[string]interface{}{
		"new_ayahs_count": input.NewAyahsCount,
		"reviewed_count":  input.ReviewedCount,
		"goal_target":     input.GoalTarget,
		"is_goal_achieved": input.IsGoalAchieved,
	}

	if err := config.DB.
		Model(&progress).
		Updates(updates).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update daily progress.",
		})
	}

	if err := config.DB.
		Where("user_id = ? AND date = ?", userID, date).
		First(&progress).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve updated daily progress.",
		})
	}

	return c.JSON(fiber.Map{
		"message":  "Daily progress updated successfully.",
		"progress": progress,
	})
}

func DeleteDailyProgress(c fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required.",
		})
	}

	date, err := parseDate(c.Params("date"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid date. Use YYYY-MM-DD format.",
		})
	}

	result := config.DB.
		Where("user_id = ? AND date = ?", userID, date).
		Delete(&models.DailyProgress{})

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete daily progress.",
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Daily progress not found.",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Daily progress deleted successfully.",
	})
}

func validateDailyProgress(
	newAyahsCount int,
	reviewedCount int,
	goalTarget int,
) error {
	if newAyahsCount < 0 {
		return errors.New("new_ayahs_count cannot be negative")
	}

	if newAyahsCount > maxDailyAyahs {
		return errors.New("new_ayahs_count cannot exceed 1000")
	}

	if reviewedCount < 0 {
		return errors.New("reviewed_count cannot be negative")
	}

	if reviewedCount > maxDailyAyahs {
		return errors.New("reviewed_count cannot exceed 1000")
	}

	if goalTarget <= 0 {
		return errors.New("goal_target must be greater than zero")
	}

	if goalTarget > maxDailyGoalTarget {
		return errors.New("goal_target cannot exceed 1000")
	}

	return nil
}

func parseDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)

	return time.Parse(dateLayout, value)
}

func parseDailyPagination(c fiber.Ctx) (int, int, error) {
	const (
		defaultPage  = 1
		defaultLimit = 30
		maxLimit     = 100
	)

	page := defaultPage
	limit := defaultLimit

	if value := c.Query("page"); value != "" {
		parsed, err := strconv.Atoi(value)

		if err != nil || parsed < 1 {
			return 0, 0, errors.New(
				"page must be a positive integer",
			)
		}

		page = parsed
	}

	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)

		if err != nil || parsed < 1 {
			return 0, 0, errors.New(
				"limit must be a positive integer",
			)
		}

		if parsed > maxLimit {
			return 0, 0, errors.New(
				"limit cannot exceed 100",
			)
		}

		limit = parsed
	}

	return page, limit, nil
}