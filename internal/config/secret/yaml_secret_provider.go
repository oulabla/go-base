package secret

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog/log"
	"go.yaml.in/yaml/v2"
)

// Secrets структура, соответствующая вашему YAML
type YAMLSecretProvider struct {
	Credentials map[string]map[string]string `yaml:"credentials"`
}

func NewYAMLSecretProvider(path string) (*YAMLSecretProvider, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			log.Error().Err(err).Msg("closing secret yaml")
		}
	}()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл: %w", err)
	}

	var s YAMLSecretProvider
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("ошибка разбора YAML: %w", err)
	}

	log.Info().Msg(fmt.Sprintf("secrets has been loaded from %s", path))

	return &s, nil
}

// Get возвращает значение секрета или пустую строку, если не найдено
func (s *YAMLSecretProvider) Get(category, key string) string {
	if s.Credentials == nil {
		return ""
	}
	if cat, ok := s.Credentials[category]; ok {
		if val, ok := cat[key]; ok {
			return val
		}
	}
	return ""
}
