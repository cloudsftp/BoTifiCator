package main

import (
	"dagger/bo-tifi-cator/internal/dagger"

	"context"
)

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

	_, err = b.Deploy(ctx, host, username, key)
	if err != nil {
		return err
	}

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
