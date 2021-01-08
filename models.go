package quartz

import (
	"errors"
	"time"
)

// Config ...
type Config struct {
	Port        int
	DatabaseURL string
}

// ErrEntityNotFound ...
var (
	ErrEntityNotFound = errors.New("Entity not found in database")
)

// Model ...
type Model struct {
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// Cron ...
type Cron struct {
	ID         string `json:"id" db:"id"`
	Expression string `json:"pattern" db:"expression"`
}

// Job ...
type Job struct {
	Model
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Timezone    string `json:"timezone" db:"timezone"`
	Schedule    []Cron `json:"schedule"`
	ContainerID string `json:"containerId" db:"container_id"`
}

// JobService ...
type JobService interface {
	Job(jobID string) (Job, error)
	Jobs() ([]Job, error)
	CreateJob(j Job) (string, error)
	DeleteJob(jobID string) error
}

// ContainerService ...
type ContainerService interface {
	BuildImage(imageTag, contextDir string) error
	RemoveImage(imageTag string) error
	Create(name, imageTag string) error
	Start(name string) error
	Stop(name string) error
	Delete(name string) error
}
