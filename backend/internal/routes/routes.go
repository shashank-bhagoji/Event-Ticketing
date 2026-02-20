package routes

import (
	"backend/internal/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Implement default CORS allowing all origins
	r.Use(cors.Default())

	r.POST("/events", handlers.CreateEvent)
	r.GET("/events", handlers.GetEvents)
	r.DELETE("/events/:id", handlers.DeleteEvent)
	r.POST("/register", handlers.RegisterForEvent)
	r.POST("/auth/login", handlers.AuthLogin)

	return r
}