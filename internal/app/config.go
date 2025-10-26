package app

type Config struct {
	App Application `yaml:"app"`
}

type Application struct{}
