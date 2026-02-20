package handlers

import (
	"net/http"
	"time"
	"backend/internal/database"
	"backend/internal/models"
	"github.com/gin-gonic/gin"
)

func CreateEvent(c *gin.Context) {
	var event models.Event

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event.AvailableCapacity = event.TotalCapacity

	database.DB.Create(&event)

	c.JSON(http.StatusOK, event)
}

func GetEvents(c *gin.Context) {
	var events []models.Event
	// Filter out expired events
	database.DB.Where("event_date > ?", time.Now()).Find(&events)
	c.JSON(http.StatusOK, events)
}

func DeleteEvent(c *gin.Context) {
	id := c.Param("id")
	
	result := database.DB.Delete(&models.Event{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}