package stage

import (
	"os"
	"strings"
)

const (
	envStageKey = "STAGE"

	envProd  = "prod"
	envDev   = "dev"
	envLocal = "local"
)

func IsProd() bool {
	return strings.ToLower(os.Getenv(envStageKey)) == envProd
}

func IsDev() bool {
	return strings.ToLower(os.Getenv(envStageKey)) == envDev
}

func IsLocal() bool {
	return strings.ToLower(os.Getenv(envStageKey)) == envLocal
}
