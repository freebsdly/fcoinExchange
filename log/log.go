package log

import (
	"fcoinExchange/conf"
	"strings"

	"go.uber.org/zap"
)

var (
	Logger *zap.SugaredLogger
)

func Init() {
	var (
		cfg    zap.Config
		logger *zap.Logger
	)

	cfg.OutputPaths = []string{conf.GetConfiguration().LogFile, "stdout"}
	switch strings.ToUpper(conf.GetConfiguration().LogLevel) {
	case "DEBUG":
		cfg = zap.NewDevelopmentConfig()
		cfg.Level.SetLevel(zap.DebugLevel)
	case "INFO":
		cfg = zap.NewProductionConfig()
		cfg.Level.SetLevel(zap.InfoLevel)
	case "WARNNING":
		cfg = zap.NewProductionConfig()
		cfg.Level.SetLevel(zap.WarnLevel)
	case "ERROR":
		cfg = zap.NewProductionConfig()
		cfg.Level.SetLevel(zap.ErrorLevel)
	}

	logger, _ = cfg.Build()
	Logger = logger.Sugar()
}
