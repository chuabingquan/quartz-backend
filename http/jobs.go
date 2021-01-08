package http

import (
	"quartz"

	"github.com/gin-gonic/gin"
)

type jobsHandler struct {
	Router     *gin.Engine
	JobService quartz.JobService
}

func (jh *jobsHandler) register() {
	v0 := jh.Router.Group("/api/v0/jobs")

	v0.GET("/")
	v0.GET("/:jobID")
	v0.POST("/")
	v0.DELETE("/:jobID")
}
