package main

import (
	"dagger/bo-tifi-cator/internal/dagger"

	"time"
)

type SSH struct {
	BaseContainer *dagger.Container
	Destination   string
	Key           *dagger.Secret
}

func NewSSH(
	// Destination to connect to (SSH destination)
	destination string,
	// Private key to connect
	key *dagger.Secret,
) *SSH {
	container := dag.
		Container().
		From("alpine:3").
		WithExec([]string{
			"apk", "add", "--no-cache", "openssh-client",
		}).
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithMountedSecret("key", key)

	return &SSH{
		Destination:   destination,
		BaseContainer: container,
	}
}

func (s *SSH) Command(command ...string) *dagger.Container {
	entrypoint := append([]string{
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-i", "key",
		s.Destination,
	}, command...)

	return s.BaseContainer.WithExec(entrypoint)
}
