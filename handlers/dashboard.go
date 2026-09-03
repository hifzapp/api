package handlers

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/mohammed-ayoub-dz/hifz/config"
	"github.com/mohammed-ayoub-dz/hifz/models"
	"github.com/mohammed-ayoub-dz/hifz/utils"
)

func GetDashboard(c fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required.",
		})
	}

	var user models.User

	if err := config.DB.
		Where("id = ?", userID).
		First(&user).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user.",
		})
	}

	var progress []models.HifzProgress

	if err := config.DB.
		Where("user_id = ?", userID).
		Order("surah_number ASC, start_ayah ASC").
		Find(&progress).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch hifz progress.",
		})
	}

	var todayProgress models.DailyProgress

	today := time.Now().Format("2006-01-02")

	err = config.DB.
		Where("user_id = ? AND date = ?", userID, today).
		First(&todayProgress).Error

	if err != nil {
		todayProgress = models.DailyProgress{
			UserID:        userID,
			Date:          time.Now(),
			NewAyahsCount: 0,
			ReviewedCount: 0,
			GoalTarget:    user.DailyGoal,
			IsGoalAchieved: false,
		}
	}

	var sessions []models.HifzSession

	if err := config.DB.
		Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Limit(10).
		Find(&sessions).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch sessions.",
		})
	}

	completed := todayProgress.NewAyahsCount

	goal := todayProgress.GoalTarget

	if goal <= 0 {
		goal = user.DailyGoal
	}

	percentage := 0

	if goal > 0 {
		percentage = int(float64(completed) / float64(goal) * 100)

		if percentage > 100 {
			percentage = 100
		}
	}

	return c.JSON(fiber.Map{
		"user": user,

		"daily": fiber.Map{
			"date":          todayProgress.Date,
			"goal":          goal,
			"completed":     completed,
			"remaining":     max(goal-completed, 0),
			"percentage":    percentage,
			"goal_achieved": todayProgress.IsGoalAchieved,
		},

		"progress": progress,

		"sessions": sessions,
	})
}

