package secret

import (
	"os"
)

type envProvider struct{}

func (e envProvider) Get(key string) string {
	return os.Getenv(key)
}
