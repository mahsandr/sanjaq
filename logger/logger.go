//nolint
package logger

import (
	"io"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var LoggerLevel = zap.NewAtomicLevelAt(zap.DebugLevel)

func GetLoggerOptions() []zap.Option {
	opts := make([]zap.Option, 4)
	opts[0] = zap.ErrorOutput(os.Stderr)
	opts[1] = zap.AddCaller()
	opts[2] = zap.AddStacktrace(zap.ErrorLevel)
	opts[3] = zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewSampler(core, time.Second, 100, 100)
	})
	return opts
}

func InitLog() *zap.Logger {
	writer := io.MultiWriter(os.Stdout)

	w := zapcore.AddSync(writer)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(
			zapcore.EncoderConfig{
				TimeKey:       "ts",
				LevelKey:      "level",
				NameKey:       "logger",
				CallerKey:     "caller",
				MessageKey:    "msg",
				StacktraceKey: "stacktrace",
				LineEnding:    zapcore.DefaultLineEnding,
				EncodeLevel:   zapcore.LowercaseLevelEncoder,
				EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
					enc.AppendString(t.Format("01-02 15:04:05 Z0700"))
				},
				EncodeDuration: zapcore.SecondsDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			},
		),
		w,
		LoggerLevel,
	)
	log := zap.New(core)
	log = log.WithOptions(GetLoggerOptions()...)
	return log
}
