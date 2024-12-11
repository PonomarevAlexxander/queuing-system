package logger

import (
	"fmt"

	"github.com/PonomarevAlexxander/queuing-system/utils/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func InitZapWrapper(log *zap.Logger) *Logger {
	return &Logger{
		Logger: log,
	}
}

func (l *Logger) Debugf(msg string, fields ...any) {
	l.Debug(fmt.Sprintf(msg, fields...))
}

func (l *Logger) Infof(msg string, fields ...any) {
	l.Info(fmt.Sprintf(msg, fields...))
}

func (l *Logger) Warnf(msg string, fields ...any) {
	l.Warn(fmt.Sprintf(msg, fields...))
}

func (l *Logger) Errorf(msg string, fields ...any) {
	l.Error(fmt.Sprintf(msg, fields...))
}

func (l *Logger) LogPanic() {
	if err := recover(); err != nil {
		l.Error("Panic captured!", zap.Any("error", err))
	}
}

func InitZapLogger(cfg config.LoggerConfig) (*zap.Logger, error) {
	level, err := zap.ParseAtomicLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level from config: %w", err)
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	config := zap.Config{
		Level:             level,
		Development:       false,
		DisableCaller:     true,
		DisableStacktrace: !cfg.Stacktrace,
		Sampling:          nil,
		Encoding:          cfg.Type,
		EncoderConfig:     encoderCfg,
		OutputPaths:       cfg.Out,
		ErrorOutputPaths:  cfg.Out,
	}

	return config.Build()
}

func getLevel(level string) zapcore.Level {
	switch level {
	case "info":
		return zap.InfoLevel
	case "warning":
		return zap.WarnLevel
	case "debug":
		return zap.DebugLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}
