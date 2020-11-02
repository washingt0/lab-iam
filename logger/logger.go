package logger

import (
	"errors"
	"lab/iam/config"
	"path"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logg *zap.Logger // nolint
)

// Setup set-up the logger
func Setup() (err error) {
	var (
		c      = config.GetConfig()
		cfg    zap.Config
		levels = map[int8]zapcore.Level{
			0: zap.DebugLevel,
			1: zap.InfoLevel,
			2: zap.WarnLevel,
			3: zap.ErrorLevel,
			4: zap.DPanicLevel,
			5: zap.PanicLevel,
			6: zap.FatalLevel,
		}
		validLevel bool
		opts       []zap.Option = make([]zap.Option, 0, 1)
	)

	for k := range levels {
		if k == int8(c.LogLevel) {
			validLevel = true
		}
	}
	if !validLevel {
		return errors.New("invalid log level")
	}

	cfg = zap.NewDevelopmentConfig()
	cfg.DisableStacktrace = true
	cfg.EncoderConfig = zap.NewProductionEncoderConfig()
	cfg.Encoding = "json"

	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.MessageKey = "message"
	cfg.EncoderConfig.LevelKey = "level"
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.NameKey = "name"
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.StacktraceKey = "stack_trace"
	cfg.InitialFields = map[string]interface{}{
		"application": c.ApplicationName,
		"version":     c.Version,
	}

	cfg.OutputPaths = []string{path.Join(c.LogPath, "access.log")}
	cfg.ErrorOutputPaths = []string{path.Join(c.LogPath, "error.log")}

	if c.LogSTDOUT {
		cfg.OutputPaths = append(cfg.OutputPaths, "stdout")
		cfg.ErrorOutputPaths = append(cfg.ErrorOutputPaths, "stderr")
	}

	if c.LogLevel > 0 {
		opts = append(opts,
			zap.IncreaseLevel(zap.LevelEnablerFunc(
				func(lvl zapcore.Level) bool {
					return lvl > levels[int8(c.LogLevel)]
				},
			)),
		)
	}

	if logg, err = cfg.Build(opts...); err != nil {
		return
	}

	zap.ReplaceGlobals(logg)

	return
}

// Sync performs fs sync for zap logger
func Sync() {
	_ = logg.Sync()
}
