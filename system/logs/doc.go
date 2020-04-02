package logs

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log *zap.Logger
	cfg zap.Config
)

func init() {
	cfg := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:      "time",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			LevelKey:     "level",
			EncodeLevel:  zapcore.CapitalColorLevelEncoder,
			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
			MessageKey:   "message",
		},
	}

	Log, _ = cfg.Build()

	Log.Info("Logger system initialized!")

}

//zapcore.Levels ("debug", "info", "warn", "error", "dpanic", "panic", and "fatal").
func SetLevel(level string) {

	cfg.Level.UnmarshalText([]byte(level))
	cfg.Level.SetLevel((cfg.Level.Level()))
}
