package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"quartz"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mholt/archiver"
)

const (
	templateDir = "templates/nodejs"
)

type jobsHandler struct {
	Router           *gin.Engine
	JobService       quartz.JobService
	ContainerService quartz.ContainerService
}

func (jh *jobsHandler) register() {
	v0 := jh.Router.Group("/api/v0/jobs")

	v0.GET("/", jh.getJobs)
	v0.GET("/:jobID", jh.getJobByID)
	v0.POST("/", jh.createJob)
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

func (jh *jobsHandler) createJob(c *gin.Context) {
	requestID := uuid.New().String()
	sourceDir := "temp/" + requestID

	defer func() {
		err := os.RemoveAll(sourceDir)
		if err != nil {
			panic(err)
		}
		if err := recover(); err != nil {
			err := jh.ContainerService.Delete(requestID)
			if err != nil {
				panic(err)
			}
			err = jh.ContainerService.RemoveImage(requestID)
			if err != nil {
				panic(err)
			}
		}
	}()

	file, err := c.FormFile("file")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, standardResponse{"Error processing deployment upload"})
		return
	}

	if _, err := os.Stat("temp"); os.IsNotExist(err) {
		os.Mkdir("temp", os.ModePerm)
	}
	sourceFileDir := sourceDir + "/" + filepath.Base(file.Filename)
	os.Mkdir(sourceDir, os.ModePerm)

	err = c.SaveUploadedFile(file, sourceFileDir)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			standardResponse{"Error processing deployment upload"})
		return
	}

	err = archiver.Unarchive(sourceFileDir, sourceDir)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			standardResponse{"Error processing deployment upload"})
		return
	}

	err = os.Remove(sourceFileDir)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			standardResponse{"Error processing deployment upload"})
		return
	}

	err = copy(templateDir+"/Dockerfile", sourceDir+"/Dockerfile")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			standardResponse{"Error processing deployment upload"})
		return
	}

	err = copy(templateDir+"/entry.js", sourceDir+"/"+requestID+".js")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			standardResponse{"Error processing deployment upload"})
		return
	}

	config, err := ioutil.ReadFile(sourceDir + "/config.json")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			standardResponse{"Error processing deployment upload"})
		return
	}

	var jobConfig JobConfig
	err = json.Unmarshal(config, &jobConfig)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			standardResponse{"Error processing deployment upload"})
		return
	}

	job := quartz.Job{
		Name:        jobConfig.Name,
		Timezone:    jobConfig.Timezone,
		ContainerID: requestID,
	}

	for _, expression := range jobConfig.Schedule {
		job.Schedule = append(job.Schedule, quartz.Cron{Expression: expression})
	}

	scheduleScript := ""
	for _, cron := range job.Schedule {
		scheduleScript += "echo '" + cron.Expression + " /app/start.sh' > /etc/crontabs/root\n"
	}
	err = ioutil.WriteFile(sourceDir+"/schedule.sh", []byte(scheduleScript), 0644)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			standardResponse{"Error processing deployment upload"})
		return
	}

	startScript := "node /app/" + requestID + ".js"
	err = ioutil.WriteFile(sourceDir+"/start.sh", []byte(startScript), 0644)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			standardResponse{"Error processing deployment upload"})
		return
	}

	items, err := ioutil.ReadDir(sourceDir)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			standardResponse{"Error processing deployment upload"})
		return
	}

	fileNames := []string{}
	for _, fileName := range items {
		fileNames = append(fileNames, sourceDir+"/"+fileName.Name())
	}

	err = archiver.Archive(fileNames, sourceDir+"/"+requestID+".tar.gz")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			standardResponse{"Error processing deployment upload"})
		return
	}

	err = jh.ContainerService.BuildImage(requestID, sourceDir+"/"+requestID+".tar.gz")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			standardResponse{"Error processing deployment upload"})
		return
	}

	err = jh.ContainerService.Create(requestID, requestID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			standardResponse{"Error processing deployment upload"})
		return
	}

	err = jh.ContainerService.Start(requestID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			standardResponse{"Error processing deployment upload"})
		return
	}

	_, err = jh.JobService.CreateJob(job)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			standardResponse{"Error processing deployment upload"})
		return
	}

	c.JSON(http.StatusOK, standardResponse{"Job is successfully deployed!"})
}

func (jh *jobsHandler) deleteJob(c *gin.Context) {
	jobID := c.Param("jobID")

	job, err := jh.JobService.Job(jobID)
	if err == quartz.ErrEntityNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, resourceNotFoundResponse)
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, internalServerErrorResponse)
		return
	}

	err = jh.ContainerService.Delete(job.ContainerID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, internalServerErrorResponse)
		return
	}

	err = jh.ContainerService.RemoveImage(job.ContainerID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, internalServerErrorResponse)
		return
	}

	err = jh.JobService.DeleteJob(jobID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, internalServerErrorResponse)
		return
	}

	c.JSON(http.StatusOK, standardResponse{"Job is successfully deleted"})
}

func copy(source, dest string) error {
	toCopy, err := ioutil.ReadFile(source)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dest, toCopy, 0644)
	if err != nil {
		return err
	}

	return nil
}
