package routes

import (
	"seta/internal/app/server/middlewares"
	"seta/internal/pkg/errorHandling"
	"seta/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// SetupRouter initializes the Gin router and sets up all application routes.
func SetupRouter(db *gorm.DB, log *zerolog.Logger) *gin.Engine {
    r := gin.Default()

    // Global Middleware
    r.Use(logger.RequestLogger(log))
    r.Use(middlewares.PrometheusMiddleware())
    r.Use(errorHandling.ErrorHandler())

    // Public Routes (No Auth Required)
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))

    // API Group with Authentication Middleware
    api := r.Group("/api")
    api.Use(middlewares.AuthMiddleware())
    {
        // Register modularized routes
        RegisterTeamRoutes(api, db)
        RegisterUserRoutes(api, db)
        RegisterFolderRoutes(api, db)
        RegisterNoteRoutes(api, db)
    }

    return r
}