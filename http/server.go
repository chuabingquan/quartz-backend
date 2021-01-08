package http

import (
	"fmt"
	"quartz"

	"github.com/gin-gonic/gin"
)

// Server ...
type Server struct {
	Port             int
	Router           *gin.Engine
	JobService       quartz.JobService
	ContainerService quartz.ContainerService
}

// Start ...
func (s *Server) Start() {
	handlers := []handler{
		&jobsHandler{
			Router:           s.Router,
			JobService:       s.JobService,
			ContainerService: s.ContainerService,
		},
	}

	for _, h := range handlers {
		h.register()
	}

	s.Router.Run(fmt.Sprintf(":%d", s.Port))
}
