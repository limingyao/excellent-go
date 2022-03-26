package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
)

type message jsonmessage.JSONMessage

func (m message) String() string {
	if m.Error != nil {
		bs, _ := json.Marshal(m)
		log.Printf("%s\n", string(bs)) // TODO remove late
		return fmt.Sprintf("%s", string(bs))
	}
	if m.Progress != nil && m.ID != "" {
		return fmt.Sprintf("%s: %s %s", m.ID, m.Status, m.Progress)
	}
	if m.Progress != nil {
		return fmt.Sprintf("%s %s", m.Status, m.Progress)
	}
	if m.ID != "" {
		return fmt.Sprintf("%s: %s", m.ID, m.Status)
	}
	return fmt.Sprintf("%s", m.Status)
}

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
func (x Client) ImagePull(ctx context.Context, imageName string, opts ...Option) error {
	defaultOpts := defaultOptions
	for _, o := range opts {
		o.apply(&defaultOpts)
	}

	reader, err := x.cli.ImagePull(
		ctx,
		imageName,
		types.ImagePullOptions{RegistryAuth: defaultOpts.registryAuth},
	)
	if err != nil {
		return err
	}

	dec := json.NewDecoder(reader)
	for {
		var jm message
		if err := dec.Decode(&jm); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		log.Print(jm)
	}

	return nil
}
