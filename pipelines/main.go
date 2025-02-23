package main

import (
	"dagger/bo-tifi-cator/internal/dagger"

	"context"
)

type BoTifiCator struct{}

const (
	GoVersion     = "1.23"
	AlpineVersion = "3.21"

	serviceName = "botificator-service"
)

func (b *BoTifiCator) BuildImage(
	ctx context.Context,
	source *dagger.Directory,
) *dagger.Container {
	return b.
		buildBaseImage(source).
		WithEntrypoint([]string{"/server"})
}

func (b *BoTifiCator) buildBaseImage(
	source *dagger.Directory,
) *dagger.Container {
	executable := b.Build(source)

	return dag.Container().
		From("alpine:"+AlpineVersion).
		//WithExposedPort(6680).
		WithFile("/server", executable)
}

func (b *BoTifiCator) Build(
	source *dagger.Directory,
) *dagger.File {
	return cachedGoBuilder(source).
		WithExec([]string{"go", "build", "-o", serviceName, "./cmd/server"}).
		File(serviceName)
}

func cachedGoBuilder(
	source *dagger.Directory,
) *dagger.Container {
	return dag.Container().
		From("golang:"+GoVersion+"-alpine"+AlpineVersion).

		// Caches
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
		WithMountedCache("/go/build-cache", dag.CacheVolume("go-build")).
		WithEnvVariable("GOCACHE", "/go/build-cache").

		// Source code
		WithDirectory("/src", source).
		WithWorkdir("/src")
}

func (b *BoTifiCator) PublishAndDeploy(
	ctx context.Context,
	source *dagger.Directory,
	actor string,
	token *dagger.Secret,
	host *dagger.Secret,
	username *dagger.Secret,
	key *dagger.Secret,
) error {
	_, err := b.PublishImage(ctx, source, actor, token)
	if err != nil {
		return err
	}

	/*
		_, err = l.Deploy(ctx, host, username, key)
		if err != nil {
			return err
		}
	*/

	return nil
}

func (b *BoTifiCator) PublishImage(
	ctx context.Context,
	source *dagger.Directory,
	actor string,
	token *dagger.Secret,
) (string, error) {
	return b.
		BuildImage(ctx, source).
		WithRegistryAuth("ghcr.io", actor, token).
		Publish(ctx, "ghcr.io/cloudsftp/botificator-service:latest")
}

func (b *BoTifiCator) Deploy(
	ctx context.Context,
	host *dagger.Secret,
	username *dagger.Secret,
	key *dagger.Secret,
) (string, error) {
	usernamePlain, err := username.Plaintext(ctx)
	if err != nil {
		return "", err
	}

	hostPlain, err := host.Plaintext(ctx)
	if err != nil {
		return "", err
	}

	return NewSSH(
		usernamePlain+"@"+hostPlain,
		key,
	).
		Command("./deploy.sh").
		Stdout(ctx)
}
