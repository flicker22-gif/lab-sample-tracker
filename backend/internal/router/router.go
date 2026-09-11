package router

import (
	"net/http"

	"sample-tracker/internal/handler"
	"sample-tracker/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func New(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
	}))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	sampleH := handler.NewSampleHandler(
		service.NewSampleService(db),
		service.NewTransferService(db),
		service.NewResultService(db),
	)
	meta := handler.NewMetaHandler(db)

	api := r.Group("/api/v1")
	{
		api.GET("/users", meta.ListUsers)
		api.GET("/locations", meta.ListLocations)

		api.GET("/samples", sampleH.List)
		api.POST("/samples", sampleH.Create)
		api.GET("/samples/:id", sampleH.Detail)
		api.POST("/samples/:id/transfers", sampleH.CreateTransfer)
		api.POST("/samples/:id/results", sampleH.CreateResult)
	}
	return r
}
