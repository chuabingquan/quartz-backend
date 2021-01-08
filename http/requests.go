package http

// JobConfig ...
type JobConfig struct {
	Name     string   `json:"name"`
	Timezone string   `json:"timezone"`
	Schedule []string `json:"schedule"`
}
