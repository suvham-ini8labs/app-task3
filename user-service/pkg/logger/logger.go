package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.SugaredLogger
}

func New(level, format string) (*Logger, error) {
	var config zap.Config
	
	if format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	
	// Set log level
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}
	config.Level = zap.NewAtomicLevelAt(zapLevel)
	
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	
	return &Logger{logger.Sugar()}, nil
}

func (l *Logger) Sync() {
	_ = l.SugaredLogger.Sync()
}

func Fatal(msg string, keysAndValues ...interface{}) {
	zap.L().Sugar().Fatalw(msg, keysAndValues...)
	os.Exit(1)
}
