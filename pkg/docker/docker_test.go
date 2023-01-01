package docker_test

import (
	"context"
	"testing"

	"github.com/docker/docker/errdefs"
	"github.com/limingyao/excellent-go/pkg/docker"
	"github.com/stretchr/testify/assert"
)

func TestImagePull(t *testing.T) {
	ast := assert.New(t)
	ctx := context.Background()
	cli, err := docker.New()
	ast.Nil(err)

	err = cli.ImagePull(ctx, "ccr.ccs.tencentyun.com/central/centos:7")
	ast.Nil(err)

	err = cli.ImagePull(ctx, "ccr.ccs.tencentyun.com/central/centos:not-found")
	ast.True(errdefs.IsNotFound(err))

	err = cli.ImagePull(
		ctx,
		"ccr.ccs.tencentyun.com/central/kms-server:1113",
		docker.WithRegistryAuth("error", "error"),
	)
	ast.True(errdefs.IsSystem(err))
}

func TestImageBuild(t *testing.T) {
	ast := assert.New(t)
	ctx := context.Background()
	cli, err := docker.New()
	ast.Nil(err)

	err = cli.ImageBuild(ctx, ".", "Dockerfile", "docker-build:unittest", docker.WithCompress())
	ast.Nil(err)

	err = cli.ImageSave(ctx, "docker-build:unittest", "/tmp/1.tar")
	ast.Nil(err)

	err = cli.ImageRemove(ctx, "docker-build:unittest")
	ast.Nil(err)

	err = cli.ImageBuild(ctx, ".", "Dockerfile_no_auth", "docker-build:unittest", docker.WithCompress())
	ast.NotNil(err)

	err = cli.ImageBuild(ctx, ".", "Dockerfile_not_found", "docker-build:unittest", docker.WithCompress())
	ast.NotNil(err)
}

func TestImageTag(t *testing.T) {
	ast := assert.New(t)
	ctx := context.Background()
	cli, err := docker.New()
	ast.Nil(err)

	err = cli.ImageTag(ctx, "ccr.ccs.tencentyun.com/central/centos:7", "ccr.ccs.tencentyun.com/central/centos:7.2")
	ast.Nil(err)
}

func TestImagePush(t *testing.T) {
	ast := assert.New(t)
	ctx := context.Background()
	cli, err := docker.New()
	ast.Nil(err)

	err = cli.ImagePush(ctx, "ccr.ccs.tencentyun.com/central/centos:7.2", docker.WithRegistryAuth("u", "p"))
	t.Log(err)
	ast.Nil(err)
}
