package image

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

// Create a Docker Client
func getDockerClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return cli, nil
}

// Inspect the Image to Retrieve the Digest
func getImageDigest(cli *client.Client, imageName string) (string, error) {
	ctx := context.Background()
	imageInspect, _, err := cli.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		return "", err
	}

	// Check if RepoDigests exists and return the first one
	if len(imageInspect.RepoDigests) > 0 {
		return imageInspect.RepoDigests[0], nil
	}
	return "", fmt.Errorf("no digest found for image %s", imageName)
}
