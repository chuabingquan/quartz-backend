package main

import (
	"fmt"
	"os"
	"quartz"
	"quartz/http"
	"quartz/postgres"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	config, err := getConfig()
	if err != nil {
		panic(err)
	}

	db, err := postgres.Open(config.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	js := &postgres.JobService{DB: db}

	server := http.Server{
		Port:       config.Port,
		Router:     gin.Default(),
		JobService: js,
	}
	server.Start()
}

func getConfig() (quartz.Config, error) {
	const (
		envPort        = "PORT"
		envDatabaseURL = "DATABASE_URL"
	)

	envNames := []string{
		envPort,
		envDatabaseURL,
	}

	values := map[string]string{}
	for _, envName := range envNames {
		val, ok := os.LookupEnv(envName)
		if !ok {
			return quartz.Config{}, fmt.Errorf("\"%s\" environment variable is required but not set", envName)
		}
		values[envName] = val
	}

	return quartz.Config{
		Port:        toIntOrPanic(values[envPort]),
		DatabaseURL: values[envDatabaseURL],
	}, nil
}

func toIntOrPanic(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Errorf("Failed to parse \"%s\" to int: %w", s, err))
	}
	return val
}
