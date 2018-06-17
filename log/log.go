package log

import (
	"fcoinExchange/conf"
	"fmt"
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
	cfg = zap.NewProductionConfig()
	cfg.OutputPaths = []string{conf.GetConfiguration().LogFile, "stdout"}
	switch strings.ToUpper(conf.GetConfiguration().LogLevel) {
	case "DEBUG":
		cfg.Level.SetLevel(zap.DebugLevel)
		break
	case "INFO":
		cfg.Level.SetLevel(zap.InfoLevel)
		break
	case "WARNNING":

		cfg.Level.SetLevel(zap.WarnLevel)
		break
	case "ERROR":
		cfg.Level.SetLevel(zap.ErrorLevel)
		break
	default:
		cfg.Level.SetLevel(zap.ErrorLevel)
		break
	}

	var err error
	logger, err = cfg.Build()
	if err != nil {
		fmt.Printf("build log failed. %s\n", err)
	}
	Logger = logger.Sugar()
}
