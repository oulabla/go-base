package config

import (
	"context"
	"fmt"
	"maps"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type yamlProvider struct {
	mu     sync.RWMutex
	values map[string]entry
}

type entry struct {
	Type  string `yaml:"type"`
	Value any    `yaml:"value"`
}

type rawRoot struct {
	Config map[string]entry `yaml:"config"`
}

func NewYAMLProvider(paths ...string) (*yamlProvider, error) {
	merged := make(map[string]entry)

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read config file %s: %w", path, err)
		}

		var root rawRoot
		if err := yaml.Unmarshal(data, &root); err != nil {
			return nil, fmt.Errorf("parse yaml %s: %w", path, err)
		}

		maps.Copy(merged, root.Config)
	}

	return &yamlProvider{
		values: merged,
	}, nil
}

func (f *yamlProvider) get(key string) entry {
	f.mu.RLock()
	defer f.mu.RUnlock()

	v, ok := f.values[key]
	if !ok {
		panic(fmt.Sprintf("config key not found: %s", key))
	}
	return v
}

func envKey(key string) string {
	return "CONFIG_" + strings.ToUpper(key)
}

func getEnvOverride(key string) (string, bool) {
	env := envKey(key)
	v, ok := os.LookupEnv(env)
	return v, ok
}

func (f *yamlProvider) GetString(ctx context.Context, key string) string {
	if v, ok := getEnvOverride(key); ok {
		return v
	}

	e := f.get(key)

	if e.Type != "string" {
		panic(fmt.Sprintf("config key %s is not string (actual: %s)", key, e.Type))
	}

	return fmt.Sprintf("%v", e.Value)
}

func (f *yamlProvider) GetInt(ctx context.Context, key string) int {
	if v, ok := getEnvOverride(key); ok {
		i, err := strconv.Atoi(v)
		if err != nil {
			panic(fmt.Sprintf("invalid int ENV override for key %s", key))
		}
		return i
	}

	e := f.get(key)

	if e.Type != "int" {
		panic(fmt.Sprintf("config key %s is not int (actual: %s)", key, e.Type))
	}

	switch v := e.Value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			panic(fmt.Sprintf("invalid int value for key %s", key))
		}
		return i
	default:
		panic(fmt.Sprintf("unsupported int type for key %s", key))
	}
}

func (f *yamlProvider) GetBool(ctx context.Context, key string) bool {
	if v, ok := getEnvOverride(key); ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			panic(fmt.Sprintf("invalid bool ENV override for key %s", key))
		}
		return b
	}

	e := f.get(key)

	if e.Type != "bool" {
		panic(fmt.Sprintf("config key %s is not bool (actual: %s)", key, e.Type))
	}

	switch v := e.Value.(type) {
	case bool:
		return v
	case string:
		b, err := strconv.ParseBool(v)
		if err != nil {
			panic(fmt.Sprintf("invalid bool value for key %s", key))
		}
		return b
	default:
		panic(fmt.Sprintf("unsupported bool type for key %s", key))
	}
}

func (f *yamlProvider) GetDuration(ctx context.Context, key string) time.Duration {
	if v, ok := getEnvOverride(key); ok {
		d, err := time.ParseDuration(v)
		if err != nil {
			panic(fmt.Sprintf("invalid duration ENV override for key %s", key))
		}
		return d
	}

	e := f.get(key)

	if e.Type != "duration" {
		panic(fmt.Sprintf("config key %s is not duration (actual: %s)", key, e.Type))
	}

	str := fmt.Sprintf("%v", e.Value)

	d, err := time.ParseDuration(str)
	if err != nil {
		panic(fmt.Sprintf("invalid duration for key %s: %s", key, str))
	}

	return d
}
