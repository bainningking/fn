package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	r.Use(Logger())
	r.Use(CORS())

	api := r.Group("/api/v1")
	{
		// Agent 管理
		agents := api.Group("/agents")
		{
			handler := NewAgentHandler(db)
			agents.GET("", handler.List)
			agents.GET("/:id", handler.Get)
			agents.DELETE("/:id", handler.Delete)
		}

		// 任务管理
		tasks := api.Group("/tasks")
		{
			handler := NewTaskHandler(db)
			tasks.POST("", handler.Create)
			tasks.GET("", handler.List)
			tasks.GET("/:id", handler.Get)
		}

		// 指标查询
		metrics := api.Group("/metrics")
		{
			handler := NewMetricHandler(db)
			metrics.GET("", handler.Query)
		}
	}

	return r
}
