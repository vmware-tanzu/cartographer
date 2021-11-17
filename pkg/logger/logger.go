package logger

import (
	"fmt"
	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"strings"
)

func SetLogLevel(logLevel string) (zap.Opts, error) {
	var level zapcore.Level
	switch strings.ToUpper(logLevel) {
	case "TRACE":
		level = zapcore.Level(-2)
	case "DEBUG":
		level = zapcore.Level(-1)
	case "INFO":
		level = zapcore.Level(0)
	case "ERROR":
		level = zapcore.Level(1)
	default:
		return nil, fmt.Errorf("if present, log-level must be one of {error, info, debug, trace}")
	}

	zapOpt := zap.Options{}
	zapOpt.Level = level

	return zap.UseFlagOptions(&zapOpt), nil
}
