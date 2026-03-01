package secret

import (
	"context"

	"github.com/rs/zerolog/log"
)

type SecretProvider interface {
	Get(category, key string) string
}

var provider SecretProvider

func Get(_ context.Context, category, key string) string {
	if provider == nil {
		log.Fatal().Msg("empty secret provider")
	}
	return provider.Get(category, key)
}

func SetProvider(p SecretProvider) {
	provider = p
}
