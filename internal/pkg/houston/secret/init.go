package secret

var (
	instance SecretProvider
)

type SecretProvider interface {
	Get(key string) string
}

func init() {
	InitEnvProvider()
}

func InitEnvProvider() {
	instance = envProvider{}
}

func SetGlobal(provider SecretProvider) {
	instance = provider
}
