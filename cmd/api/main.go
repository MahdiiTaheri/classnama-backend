package main

import (
	"expvar"
	"runtime"

	"github.com/MahdiiTaheri/classnama-backend/internal/env"
	"go.uber.org/zap"
)

const version = "0.1.0"

func main() {
	// Load config
	cfg := config{
		addr: env.GetString("ADDR", ":8000"),
		env:  env.GetString("ENV", "development"),
	}

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	app := &application{
		config: cfg,
		logger: logger,
	}

	// Publish some expvar metrics
	expvar.NewString("version").Set(version)
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	// Run server
	logger.Fatal(app.run(app.mount()))
}
