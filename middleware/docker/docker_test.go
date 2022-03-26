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

	err = cli.ImagePull(ctx, "ccr.ccs.tencentyun.com/central/centos:not_found")
	ast.True(errdefs.IsNotFound(err))

	err = cli.ImagePull(
		ctx,
		"ccr.ccs.tencentyun.com/central/kms-server:1113",
		WithRegistryAuth("error", "error"),
	)
	ast.True(errdefs.IsSystem(err))
}
