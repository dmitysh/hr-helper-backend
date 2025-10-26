package config

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"hr-helper/internal/pkg/houston/stage"
)

const (
	defaultConfigPath = "configs"
	configPathEnvKey  = "CONFIG_PATH"

	prodConfigName  = "values-prod"
	devConfigName   = "values-docker"
	localConfigName = "values-local"
)

var (
	ErrUnknownStage = errors.New("unknown stage")
)

func ReadAndParseYAML(out any) error {
	if err := ReadYAML(); err != nil {
		return fmt.Errorf("can't read yaml config: %w", err)
	}
	if err := ParseYAML(out); err != nil {
		return fmt.Errorf("can't parse yaml config: %w", err)
	}

	return nil
}

func ReadYAML() error {
	cfgName := selectCfgName()
	if cfgName == "" {
		return ErrUnknownStage
	}

	cfgPath := os.Getenv(configPathEnvKey)
	if cfgPath == "" {
		cfgPath = defaultConfigPath
	}

	viper.SetConfigType("yaml")
	viper.AddConfigPath(cfgPath)
	viper.SetConfigName(cfgName)

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("can't read in yaml config: %w", err)
	}

	return nil
}

func ParseYAML(out any) error {
	cfgName := selectCfgName()
	if cfgName == "" {
		return ErrUnknownStage
	}

	cfgPath := os.Getenv(configPathEnvKey)
	if cfgPath == "" {
		cfgPath = defaultConfigPath
	}
	data, err := os.ReadFile(fmt.Sprintf("%s.%s", path.Join(cfgPath, cfgName), "yaml"))
	if err != nil {
		return fmt.Errorf("can't read config: %w", err)
	}

	if err = yaml.Unmarshal(data, out); err != nil {
		return fmt.Errorf("can't unmarshal yaml: %w", err)
	}

	return nil
}

func selectCfgName() string {
	var cfgName string
	switch {
	case stage.IsDev():
		cfgName = devConfigName
	case stage.IsProd():
		cfgName = prodConfigName
	case stage.IsLocal():
		cfgName = localConfigName
	default:
		cfgName = ""
	}

	return cfgName
}
