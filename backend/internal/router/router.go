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
	prodH := handler.NewProductionHandler(service.NewProductionService(db))
	waferH := handler.NewWaferHandler(service.NewWaferService(db))
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

		// 工艺 / 机台
		api.GET("/processes", prodH.ListProcesses)
		api.GET("/machines", prodH.ListMachines)
		api.PUT("/machines/:id/status", prodH.UpdateMachineStatus)
		api.GET("/machines/:id/status-logs", prodH.ListMachineLogs)

		// 晶圆批次与加工
		api.GET("/lots", prodH.ListLots)
		api.POST("/lots", prodH.CreateLot)
		api.GET("/lots/:id", prodH.GetLot)
		api.POST("/lots/:id/start", prodH.StartProcess)
		api.POST("/lots/:id/end", prodH.EndProcess)
		api.POST("/lots/:id/complete", prodH.CompleteLot)
		api.GET("/lots/:id/wafers", waferH.ListByLot)

		// 晶圆 bin map
		api.GET("/wafers/:id/maps", waferH.ListVersions)
		api.POST("/wafers/:id/maps", waferH.UploadMap)
		api.GET("/maps/:id", waferH.GetMap)
		api.GET("/maps/:id/download", waferH.DownloadMap)
	}
	return r
}
