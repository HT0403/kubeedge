package image

import (
	"context"
	"testing"
)

func TestGetDockerClient(t *testing.T) {
	ctx := context.TODO()
	cli, err := getDockerClient()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cli == nil {
		t.Fatalf("Expected valid Docker client, got nil")
	}

	// Optional: Check the version to verify the client is usable
	_, err = cli.ServerVersion(ctx)
	if err != nil {
		t.Fatalf("Expected no error calling ServerVersion, got %v", err)
	}
}

func TestGetImageDigest(t *testing.T) {
	cli, err := getDockerClient()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Use a known image that exists locally for testing
	imageName := "nginx:latest" // Replace with a test image in your environment
	digest, err := getImageDigest(cli, imageName)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if digest == "" {
		t.Fatalf("Expected a valid digest, got empty string")
	}

	// Optional: Validate the format of the digest (starts with imageName@sha256)
	expectedPrefix := imageName + "@sha256:"
	if len(digest) <= len(expectedPrefix) || digest[:len(expectedPrefix)] != expectedPrefix {
		t.Fatalf("Expected digest to start with %s, got %s", expectedPrefix, digest)
	}
}
