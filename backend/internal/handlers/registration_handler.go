package handlers

import (
	"errors"
	"net/http"
	"backend/internal/database"
	"backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func RegisterForEvent(c *gin.Context) {
	var input struct {
		UserID  uint `json:"user_id"`
		EventID uint `json:"event_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {

		// Check if the user is already registered for this event
		var existingReg models.Registration
		if err := tx.Where("user_id = ? AND event_id = ?", input.UserID, input.EventID).First(&existingReg).Error; err == nil {
			return errors.New("already_registered")
		}

		var event models.Event

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&event, input.EventID).Error; err != nil {
			return err
		}

		if event.AvailableCapacity <= 0 {
			return errors.New("no_slots")
		}

		event.AvailableCapacity -= 1
		tx.Save(&event)

		reg := models.Registration{
			UserID:  input.UserID,
			EventID: input.EventID,
			Status:  "registered",
		}

		tx.Create(&reg)

		return nil
	})

	if err != nil {
		if err.Error() == "already_registered" {
			c.JSON(http.StatusConflict, gin.H{"error": "You are already registered for this event."})
			return
		}
		if err.Error() == "no_slots" {
			c.JSON(http.StatusConflict, gin.H{"error": "No available slots for this event."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Registered successfully"})
}