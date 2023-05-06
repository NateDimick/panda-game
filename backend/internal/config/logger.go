package config

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func LoggerConfig() zap.Config {
	return zap.Config{
		Level:    zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:    "log",
			LevelKey:      "level",
			EncodeLevel:   zapcore.CapitalLevelEncoder,
			TimeKey:       "when",
			EncodeTime:    zapcore.RFC3339TimeEncoder,
			CallerKey:     "where",
			EncodeCaller:  zapcore.ShortCallerEncoder,
			FunctionKey:   "func",
			StacktraceKey: "trace",
			LineEnding:    "\n",
		},
		Sampling:         nil,
		OutputPaths:      []string{"panda-server.log", "stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}
