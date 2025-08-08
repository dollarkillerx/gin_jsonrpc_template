package logs

import (
	"os"
	"path/filepath"
	"time"

	"github.com/google/feitian/pkg/common/config"
	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitLog(loggerConfig config.LoggerConfig) {
	if loggerConfig.Filename != "" {
		_ = os.MkdirAll(filepath.Dir(loggerConfig.Filename), 0o755)
	}

	rotatingLogger := &lumberjack.Logger{
		Filename:   loggerConfig.Filename,
		MaxSize:    loggerConfig.MaxSize,
		MaxBackups: 1,
		MaxAge:     28,
		Compress:   true,
	}

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	multi := zerolog.MultiLevelWriter(consoleWriter, rotatingLogger)
	log.Logger = zerolog.New(multi).With().Caller().Timestamp().Logger()
	log.Info().Msg("Logger initialized")
}
