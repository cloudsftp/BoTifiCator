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

// Builds the service and runs all tests (none right now)
func (b *BoTifiCator) BuildAndTestAll(
	ctx context.Context,
	source *dagger.Directory,
) (string, error) {
	_, err := b.Lint(ctx, source)
	if err != nil {
		return "", err
	}

	b.Build(source)

	_, err = b.Test(ctx, source)
	if err != nil {
		return "", err
	}

	b.BuildImage(ctx, source)

	/*
		_, err := b.TestIntegration(ctx, source, mittlifeSource)
		if err != nil {
			return "", err
		}
	*/

	output := "SUCCESS"
	return output, nil
}

// Runs a linter
func (b *BoTifiCator) Lint(ctx context.Context, source *dagger.Directory) (string, error) {
	return cachedGoBuilder(source).
		WithExec([]string{"golangci-lint", "run"}).
		Stdout(ctx)
}

// Builds the service executable
func (b *BoTifiCator) Build(
	source *dagger.Directory,
) *dagger.File {
	return cachedGoBuilder(source).
		WithExec([]string{"go", "build", "-o", serviceName, "./cmd/server"}).
		File(serviceName)
}

// Runs unit tests
func (b *BoTifiCator) Test(
	ctx context.Context,
	source *dagger.Directory,
) (string, error) {
	return cachedGoBuilder(source).
		WithExec([]string{"go", "test", "./..."}).
		Stdout(ctx)
}

const golangciLintURL = "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

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

		// Linter
		WithExec([]string{"go", "install", golangciLintURL}).

		// Source code
		WithDirectory("/src", source).
		WithWorkdir("/src")
}
