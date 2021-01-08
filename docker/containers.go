package docker

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// ContainerService ...
type ContainerService struct {
	docker *client.Client
}

// NewContainerService ...
func NewContainerService() (*ContainerService, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("Failed to initialise container service: %w", err)
	}
	return &ContainerService{
		docker: cli,
	}, nil
}

// Close ...
func (cs *ContainerService) Close() error {
	return cs.docker.Close()
}

// BuildImage ...
func (cs *ContainerService) BuildImage(imageTag, contextDir string) error {
	ctx := context.Background()

	buildCtx, err := os.Open(contextDir)
	if err != nil {
		return fmt.Errorf("Failed to build docker image: %w", err)
	}
	defer buildCtx.Close()

	opts := types.ImageBuildOptions{
		SuppressOutput: false,
		Tags:           []string{imageTag},
		Dockerfile:     "Dockerfile",
		// BuildArgs:      args,
	}

	resp, err := cs.docker.ImageBuild(ctx, buildCtx, opts)
	if err != nil {
		return fmt.Errorf("Failed to build docker image: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)

	return nil
}

// RemoveImage ...
func (cs *ContainerService) RemoveImage(imageTag string) error {
	ctx := context.Background()
	_, err := cs.docker.ImageRemove(ctx, imageTag, types.ImageRemoveOptions{
		Force: true,
	})
	if err != nil {
		return fmt.Errorf("Failed to remove docker image: %w", err)
	}
	return nil
}

// Create ...
func (cs *ContainerService) Create(name, imageTag string) error {
	ctx := context.Background()
	_, err := cs.docker.ContainerCreate(ctx, &container.Config{
		Image: imageTag,
	}, &container.HostConfig{
		AutoRemove: true,
	}, nil, nil, imageTag)
	if err != nil {
		return fmt.Errorf("Failed to create container from image: %w", err)
	}
	return nil
}

// Start ...
func (cs *ContainerService) Start(name string) error {
	ctx := context.Background()
	err := cs.docker.ContainerStart(ctx, name, types.ContainerStartOptions{})
	if err != nil {
		return fmt.Errorf("Failed to start container: %w", err)
	}
	return nil
}

// Stop ...
func (cs *ContainerService) Stop(name string) error {
	ctx := context.Background()
	err := cs.docker.ContainerStop(ctx, name, nil)
	if err != nil {
		return fmt.Errorf("Failed to stop container: %w", err)
	}
	return nil
}

// Delete ...
func (cs *ContainerService) Delete(name string) error {
	ctx := context.Background()
	err := cs.docker.ContainerRemove(ctx, name, types.ContainerRemoveOptions{
		Force: true,
	})
	if err != nil {
		return fmt.Errorf("Failed to delete container: %w", err)
	}
	return nil
}
