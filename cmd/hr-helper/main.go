package main

import (
	"context"

	"hr-helper/internal/app"
	"hr-helper/internal/pkg/houston/closer"
	"hr-helper/internal/pkg/houston/config"
	"hr-helper/internal/pkg/houston/loggy"
	"hr-helper/internal/pkg/houston/secret"
)

func main() {
	ctx, appCancel := context.WithCancel(context.Background())

	loggy.InitDefault()
	defer loggy.Sync()
	defer appCancel()

	cfg := app.Config{}
	err := config.ReadAndParseYAML(&cfg)
	if err != nil {
		loggy.Fatal("can't read and parse config:", err)
	}

	secret.InitEnvProvider()

	closer.SetShutdownTimeout(config.Duration("app.graceful_shutdown_timeout"))

	a := app.NewApp(cfg)
	a.Run(ctx)
}
