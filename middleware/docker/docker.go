package docker

import (
	"context"

	"github.com/docker/cli/cli/command/image/build"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/docker/pkg/progress"
	"github.com/pkg/errors"
)

type Client struct {
	cli *client.Client
}

func New() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Client{cli: cli}, nil
}

// ImagePull Pull an image, like docker pull
// [reference] github.com/docker/cli/cli/command/image/pull.go
func (x Client) ImagePull(ctx context.Context, imageName string, opts ...Option) error {
	defaultOpts := defaultOptions
	for _, o := range opts {
		o.apply(&defaultOpts)
	}

	responseBody, err := x.cli.ImagePull(
		ctx,
		imageName,
		types.ImagePullOptions{RegistryAuth: defaultOpts.registryAuth},
	)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	if err := printStream(responseBody); err != nil {
		return err
	}

	return nil
}

// ImageBuild build an image, like docker build
// tag: Name and optionally a tag in the 'name:tag' format
// [reference] github.com/docker/cli/cli/command/image/build.go
func (x Client) ImageBuild(ctx context.Context, contextDir, dockerfileName, tag string, opts ...Option) error {
	defaultOpts := defaultOptions
	for _, o := range opts {
		o.apply(&defaultOpts)
	}

	contextDir, relDockerfile, err := build.GetContextFromLocalDir(contextDir, dockerfileName)
	if err != nil {
		return errors.Errorf("unable to prepare context: %s", err)
	}

	excludes, err := build.ReadDockerignore(contextDir)
	if err != nil {
		return err
	}

	if err := build.ValidateContextDirectory(contextDir, excludes); err != nil {
		return errors.Errorf("error checking context: %s", err)
	}

	// and canonicalize dockerfile name to a platform-independent one
	relDockerfile = archive.CanonicalTarNameForPath(relDockerfile)

	excludes = build.TrimBuildFilesFromExcludes(excludes, relDockerfile, false)
	buildCtx, err := archive.TarWithOptions(contextDir, &archive.TarOptions{
		ExcludePatterns: excludes,
		ChownOpts:       &idtools.Identity{UID: 0, GID: 0},
	})
	if err != nil {
		return err
	}

	if defaultOpts.compress {
		if buildCtx, err = build.Compress(buildCtx); err != nil {
			return err
		}
	}

	// setup an upload progress bar
	body := progress.NewProgressReader(
		buildCtx,
		&progressLog{sf: &rawProgressFormatter{}},
		0,
		"",
		"Sending build context to Docker daemon",
	)

	buildOptions := types.ImageBuildOptions{
		Tags:       []string{tag},
		Dockerfile: relDockerfile,
	}

	response, err := x.cli.ImageBuild(ctx, body, buildOptions)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if err := printStream(response.Body); err != nil {
		return err
	}

	return nil
}

// ImageSave save an image, like docker save
func (x Client) ImageSave(ctx context.Context) error {
	return nil
}

// ImageRemove remove an image, like docker rmi
func (x Client) ImageRemove(ctx context.Context) error {
	return nil
}

// RunContainer run container in the background, like docker run -d ...
func (x Client) RunContainer(ctx context.Context) error {
	return nil
}

// ContainerStop stop container, like docker stop
func (x Client) ContainerStop(ctx context.Context) error {
	return nil
}

// ContainerRemove remove container, like docker rm
func (x Client) ContainerRemove(ctx context.Context) error {
	return nil
}
