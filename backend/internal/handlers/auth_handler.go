package handlers

import (
	"backend/internal/database"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthLogin(c *gin.Context) {
	var body struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		IsOrganizer bool   `json:"is_organizer"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	result := database.DB.Where("email = ?", body.Email).First(&user)

	if result.Error == nil {
		// User exists
		if user.Password != body.Password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
			return
		}
		
		// Ensure role matches their selection
		expectedRole := "user"
		if body.IsOrganizer {
			expectedRole = "organizer"
		}
		
		if user.Role != expectedRole {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid role for this user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user": user})
		return
	}

	// User does not exist, check if creating an organizer
	targetRole := "user"
	if body.IsOrganizer {
		targetRole = "organizer"
		// Check if an organizer already exists
		var orgCount int64
		database.DB.Model(&models.User{}).Where("role = ?", "organizer").Count(&orgCount)
		if orgCount > 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "An organizer already exists."})
			return
		}
	}

	// Create user
	newUser := models.User{
		Name:     body.Email, // Using email as name for simplicity since there's no name input
		Email:    body.Email,
		Password: body.Password,
		Role:     targetRole,
	}

	if err := database.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully", "user": newUser})
}
