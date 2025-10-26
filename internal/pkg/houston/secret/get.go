package secret

func GetString(key string) string {
	return instance.Get(key)
}
