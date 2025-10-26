package config

import (
	"time"

	"github.com/spf13/viper"

	"hr-helper/internal/pkg/houston/loggy"
)

func String(key string) string {
	val := viper.GetString(key)
	if val == "" {
		loggy.Warnln("config value for key", key, "is empty")
	}

	return val
}

func Bool(key string) bool {
	return viper.GetBool(key)
}

func Int(key string) int {
	return viper.GetInt(key)
}

func Float(key string) float64 {
	return viper.GetFloat64(key)
}

func Duration(key string) time.Duration {
	return viper.GetDuration(key)
}
