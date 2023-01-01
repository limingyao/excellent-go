package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/cli/cli/command/image/build"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
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

	responseBody, err := x.cli.ImagePull(ctx, imageName, types.ImagePullOptions{
		RegistryAuth: defaultOpts.registryAuth,
	})
	if err != nil {
		return err
	}
	defer responseBody.Close()

	if err := decodeStream(responseBody); err != nil {
		return err
	}

	return nil
}

// ImageBuild build an image, like docker build
// tag: Name and optionally a tag in the 'name:tag' format
// [reference] github.com/docker/cli/cli/command/image/build.go
func (x Client) ImageBuild(ctx context.Context, contextDir, dockerfileName, imageName string, opts ...Option) error {
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

	// 获取文件夹大小
	if _, err := os.Stat(contextDir); !(err == nil || os.IsExist(err)) {
		return fmt.Errorf("%s not exist", contextDir)
	}
	var totalSize int64
	if err := filepath.Walk(contextDir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return err
	}); err != nil {
		return err
	}

	// and canonicalize dockerfile name to a platform-independent one
	relDockerfile = archive.CanonicalTarNameForPath(relDockerfile)

	excludes = build.TrimBuildFilesFromExcludes(excludes, relDockerfile, false)
	reader, err := archive.TarWithOptions(contextDir, &archive.TarOptions{
		ExcludePatterns: excludes,
		ChownOpts:       &idtools.Identity{UID: 0, GID: 0},
	})
	if err != nil {
		return err
	}

	if defaultOpts.compress {
		if reader, err = build.Compress(reader); err != nil {
			return err
		}
	}

	// setup an upload progress bar
	progressReader := progress.NewProgressReader(
		reader,
		&progressLog{sf: &rawProgressFormatter{}},
		totalSize,
		"",
		"Sending build context to Docker daemon",
	)

	buildOptions := types.ImageBuildOptions{
		Tags:       []string{imageName},
		Dockerfile: relDockerfile,
	}

	response, err := x.cli.ImageBuild(ctx, progressReader, buildOptions)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if err := decodeStream(response.Body); err != nil {
		return err
	}

	return nil
}

// ImageSave save an image, like docker save
func (x Client) ImageSave(ctx context.Context, imageName, output string) error {
	return x.ImagesSave(ctx, []string{imageName}, output)
}

// ImagesSave save images, like docker save
func (x Client) ImagesSave(ctx context.Context, imageNames []string, output string) error {
	outputFile, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil
	}
	defer outputFile.Close()

	reader, err := x.cli.ImageSave(ctx, imageNames) // 这里会占用一定的耗时
	if err != nil {
		return err
	}
	defer reader.Close()

	var totalSize int64 = 0
	for _, imageName := range imageNames {
		inspectResp, err := x.ImageInspect(ctx, imageName)
		if err != nil {
			return err
		}
		totalSize += inspectResp.Size
	}

	// setup a save progress bar
	progressReader := progress.NewProgressReader(
		reader,
		&progressLog{sf: &rawProgressFormatter{}},
		totalSize,
		"",
		"Downloading from Docker daemon",
	)

	_, err = io.Copy(outputFile, progressReader)
	return err
}

// ImageRemove remove an image, like docker rmi
func (x Client) ImageRemove(ctx context.Context, imageName string) error {
	_, err := x.cli.ImageRemove(ctx, imageName, types.ImageRemoveOptions{})
	if err != nil && !errdefs.IsNotFound(err) {
		return err
	}
	return nil
}

// ImageInspect inspect image, like docker inspect imageId ...
func (x Client) ImageInspect(ctx context.Context, imageName string) (*types.ImageInspect, error) {
	inspectResp, _, err := x.cli.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		return nil, err
	}
	return &inspectResp, nil
}

// ImagePush Push an image, like docker push
func (x Client) ImagePush(ctx context.Context, imageName string, opts ...Option) error {
	defaultOpts := defaultOptions
	for _, o := range opts {
		o.apply(&defaultOpts)
	}
	responseBody, err := x.cli.ImagePush(ctx, imageName, types.ImagePushOptions{
		RegistryAuth: defaultOpts.registryAuth,
	})
	if err != nil {
		return err
	}
	defer responseBody.Close()

	if err := decodeStream(responseBody); err != nil {
		return err
	}

	return nil
}

// ImageTag Push an image, like docker tag
func (x Client) ImageTag(ctx context.Context, sourceTag, targetTag string) error {
	return x.cli.ImageTag(ctx, sourceTag, targetTag)
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
