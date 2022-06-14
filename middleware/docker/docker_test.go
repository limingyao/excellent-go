package docker

import (
	"context"
	"testing"

	"github.com/docker/docker/errdefs"
	"github.com/stretchr/testify/assert"
)

func TestImagePull(t *testing.T) {
	ast := assert.New(t)
	ctx := context.Background()
	cli, err := New()
	ast.Nil(err)

	err = cli.ImagePull(ctx, "ccr.ccs.tencentyun.com/central/centos:7")
	ast.Nil(err)

	err = cli.ImagePull(ctx, "ccr.ccs.tencentyun.com/central/centos:not-found")
	ast.True(errdefs.IsNotFound(err))

	err = cli.ImagePull(
		ctx,
		"ccr.ccs.tencentyun.com/central/kms-server:1113",
		WithRegistryAuth("error", "error"),
	)
	ast.True(errdefs.IsSystem(err))
}

func TestImageBuild(t *testing.T) {
	ast := assert.New(t)
	ctx := context.Background()
	cli, err := New()
	ast.Nil(err)

	err = cli.ImageBuild(ctx, ".", "Dockerfile", "docker-build:unittest", WithCompress())
	ast.Nil(err)

	err = cli.ImageSave(ctx, "docker-build:unittest", "/tmp/1.tar")
	ast.Nil(err)

	err = cli.ImageRemove(ctx, "docker-build:unittest")
	ast.Nil(err)

	err = cli.ImageBuild(ctx, ".", "Dockerfile_no_auth", "docker-build:unittest", WithCompress())
	ast.NotNil(err)

	err = cli.ImageBuild(ctx, ".", "Dockerfile_not_found", "docker-build:unittest", WithCompress())
	ast.NotNil(err)
}
