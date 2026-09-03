package handlers

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/mohammed-ayoub-dz/hifz/config"
	"github.com/mohammed-ayoub-dz/hifz/models"
	"github.com/mohammed-ayoub-dz/hifz/utils"
	"gorm.io/gorm"
)

type CreateHifzProgressInput struct {
	SurahNumber  int  `json:"surah_number"`
	StartAyah    int  `json:"start_ayah"`
	EndAyah      int  `json:"end_ayah"`
	MasteryLevel int  `json:"mastery_level"`
	Repetitions  int  `json:"repetitions"`
}

type UpdateHifzProgressInput struct {
	SurahNumber  int  `json:"surah_number"`
	StartAyah    int  `json:"start_ayah"`
	EndAyah      int  `json:"end_ayah"`
	MasteryLevel int  `json:"mastery_level"`
	Repetitions  int  `json:"repetitions"`
}

func CreateHifzProgress(c fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required.",
		})
	}

	var input CreateHifzProgressInput

	if err := c.Bind().JSON(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body.",
		})
	}

	if err := validateHifzProgress(
		input.SurahNumber,
		input.StartAyah,
		input.EndAyah,
		input.MasteryLevel,
		input.Repetitions,
	); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var existing models.HifzProgress

	err = config.DB.
		Where(`
			user_id = ?
			AND surah_number = ?
			AND start_ayah <= ?
			AND end_ayah >= ?
		`,
			userID,
			input.SurahNumber,
			input.EndAyah,
			input.StartAyah,
		).
		First(&existing).Error

	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "This ayah range overlaps with an existing progress record.",
		})
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check existing progress.",
		})
	}

	progress := models.HifzProgress{
		UserID:        userID,
		SurahNumber:   input.SurahNumber,
		StartAyah:     input.StartAyah,
		EndAyah:       input.EndAyah,
		MasteryLevel:  input.MasteryLevel,
		Repetitions:   input.Repetitions,
	}

	if err := config.DB.Create(&progress).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create hifz progress.",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "Hifz progress created successfully.",
		"progress": progress,
	})
}

func GetUserHifzProgress(c fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required.",
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

	return c.JSON(fiber.Map{
		"progress": progress,
	})
}

func GetHifzProgressByID(c fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required.",
		})
	}

	progressID, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid progress ID.",
		})
	}

	var progress models.HifzProgress

	err = config.DB.
		Where("id = ? AND user_id = ?", progressID, userID).
		First(&progress).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Hifz progress not found.",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch hifz progress.",
		})
	}

	return c.JSON(fiber.Map{
		"progress": progress,
	})
}

func UpdateHifzProgress(c fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required.",
		})
	}

	progressID, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid progress ID.",
		})
	}

	var progress models.HifzProgress

	err = config.DB.
		Where("id = ? AND user_id = ?", progressID, userID).
		First(&progress).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Hifz progress not found.",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch hifz progress.",
		})
	}

	var input UpdateHifzProgressInput

	if err := c.Bind().JSON(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body.",
		})
	}

	if err := validateHifzProgress(
		input.SurahNumber,
		input.StartAyah,
		input.EndAyah,
		input.MasteryLevel,
		input.Repetitions,
	); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if input.SurahNumber != progress.SurahNumber ||
		input.StartAyah != progress.StartAyah ||
		input.EndAyah != progress.EndAyah {

		var existing models.HifzProgress

		err = config.DB.
			Where(`
				user_id = ?
				AND surah_number = ?
				AND id <> ?
				AND start_ayah <= ?
				AND end_ayah >= ?
			`,
				userID,
				input.SurahNumber,
				progress.ID,
				input.EndAyah,
				input.StartAyah,
			).
			First(&existing).Error

		if err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "The new ayah range overlaps with an existing progress record.",
			})
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check existing progress.",
			})
		}
	}

	updates := map[string]interface{}{
		"surah_number":  input.SurahNumber,
		"start_ayah":    input.StartAyah,
		"end_ayah":      input.EndAyah,
		"mastery_level": input.MasteryLevel,
		"repetitions":   input.Repetitions,
	}

	if err := config.DB.
		Model(&progress).
		Updates(updates).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update hifz progress.",
		})
	}

	progress.UpdatedAt = time.Now()

	if err := config.DB.
		Model(&progress).
		Update("updated_at", progress.UpdatedAt).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update hifz progress timestamp.",
		})
	}

	if err := config.DB.
		Where("id = ? AND user_id = ?", progress.ID, userID).
		First(&progress).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve updated hifz progress.",
		})
	}

	return c.JSON(fiber.Map{
		"message":  "Hifz progress updated successfully.",
		"progress": progress,
	})
}

func DeleteHifzProgress(c fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required.",
		})
	}

	progressID, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid progress ID.",
		})
	}

	result := config.DB.
		Where("id = ? AND user_id = ?", progressID, userID).
		Delete(&models.HifzProgress{})

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete hifz progress.",
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Hifz progress not found.",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Hifz progress deleted successfully.",
	})
}

func validateHifzProgress(
	surahNumber int,
	startAyah int,
	endAyah int,
	masteryLevel int,
	repetitions int,
) error {
	if surahNumber < 1 || surahNumber > 114 {
		return errors.New("surah_number must be between 1 and 114")
	}

	if startAyah < 1 {
		return errors.New("start_ayah must be greater than zero")
	}

	if endAyah < startAyah {
		return errors.New("end_ayah must be greater than or equal to start_ayah")
	}

	if masteryLevel < 1 || masteryLevel > 5 {
		return errors.New("mastery_level must be between 1 and 5")
	}

	if repetitions < 0 {
		return errors.New("repetitions cannot be negative")
	}

	return nil
}

func ParseID(value string) (uint, error) {
	id, err := strconv.ParseUint(value, 10, 64)

	if err != nil || id == 0 {
		return 0, errors.New("invalid ID")
	}

	return uint(id), nil
}


