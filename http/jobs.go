package http

import (
	"log"
	"net/http"
	"quartz"

	"github.com/gin-gonic/gin"
)

type jobsHandler struct {
	Router     *gin.Engine
	JobService quartz.JobService
}

func (jh *jobsHandler) register() {
	v0 := jh.Router.Group("/api/v0/jobs")

	v0.GET("/", jh.getJobs)
	v0.GET("/:jobID", jh.getJobByID)
	v0.POST("/")
	v0.DELETE("/:jobID", jh.deleteJob)
}

func (jh *jobsHandler) getJobs(c *gin.Context) {
	jobs, err := jh.JobService.Jobs()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			internalServerErrorResponse)
		return
	}
	c.JSON(http.StatusOK, jobs)
}

func (jh *jobsHandler) getJobByID(c *gin.Context) {
	jobID := c.Param("jobID")

	job, err := jh.JobService.Job(jobID)
	log.Println(err)
	switch err {
	case quartz.ErrEntityNotFound:
		c.AbortWithStatusJSON(http.StatusNotFound, resourceNotFoundResponse)
		break
	case nil:
		c.JSON(http.StatusOK, job)
		break
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, internalServerErrorResponse)
		break
	}
}

func (jh *jobsHandler) createJob(c *gin.Context) {}

func (jh *jobsHandler) deleteJob(c *gin.Context) {
	jobID := c.Param("jobID")

	err := jh.JobService.DeleteJob(jobID)
	switch err {
	case quartz.ErrEntityNotFound:
		c.AbortWithStatusJSON(http.StatusNotFound, resourceNotFoundResponse)
		break
	case nil:
		c.JSON(http.StatusOK, standardResponse{"Job successfully deleted"})
		break
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, internalServerErrorResponse)
		break
	}
}
