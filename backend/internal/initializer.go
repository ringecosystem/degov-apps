package internal

import (
	"log/slog"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"

	"github.com/ringecosystem/degov-apps/internal/config"
	"github.com/ringecosystem/degov-apps/internal/database"
)

func AppInit() {
	InitIDGenerator(1)
	loadDotEnv()
	initConfig()
	initLog()
	initDB()
}

func loadDotEnv() {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("No .env file found, using default environment variables")
	}
}

func initConfig() {
	err := config.InitConfig()
	if err != nil {
		slog.Error("Failed to initialize configuration", "error", err)
		panic(err)
	}
}

func initLog() {
	var zapL *zap.Logger
	var err error

	env := config.GetAppEnv()

	if env.IsDevelopment() {
		zapL, err = zap.NewDevelopment()
		slog.Info("set log mode to [development]")
	}
	if env.IsStaging() {
		zapL, err = zap.NewProduction()
		slog.Info("set log mode to [staging]")
	}
	if env.IsProduction() {
		zapConfig := zap.NewProductionConfig()
		zapConfig.Encoding = config.GetLogFormat()
		zapL, err = zapConfig.Build()
		slog.Info("set log mode to [production]")
	}

	if err != nil {
		panic(err)
	}
	defer zapL.Sync()

	zaphandler := zapslog.NewHandler(zapL.Core(), zapslog.WithCaller(true), zapslog.WithCallerSkip(1), zapslog.AddStacktraceAt(slog.LevelError))
	logger := slog.New(zaphandler)

	slog.SetDefault(logger)
}

func initDB() {
	err := database.InitDB()
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		panic(err)
	}
	slog.Info("Database initialized successfully")
}
