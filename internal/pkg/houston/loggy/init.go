package loggy

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	instance *zap.SugaredLogger

	defaultLogLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	defaultOut      = os.Stdout
)

func InitDefault() {
	SetGlobal(NewJSON(defaultOut))
	Infoln("logger successfully inited")
}

func NewJSON(out io.Writer) *zap.SugaredLogger {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}),
		zapcore.AddSync(out),
		defaultLogLevel,
	)

	return zap.New(core).Sugar()
}

func SetGlobal(logger *zap.SugaredLogger) {
	instance = logger
}

func Sync() {
	_ = instance.Sync()
}
