package log

import (
	"os"

	"github.com/kataras/golog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logFileMaxAge        = 7
	logFileMaxSize       = 500
	logFileMaxRotataSize = 500
	logFileMaxBackups    = 3
)

func NewEnvLog(msenv string) *golog.Logger {
	switch msenv {
	case "dev":
		golog.Info("setting log level to debug for dev env")
		return New("debug", "stdout")
	case "prod":
		golog.Info("setting log level to info for prod env")
		return New("info", "json")
	case "stag":
		golog.Info("setting log level to info for prod env")
		return New("info", "file")
	default:
		golog.Info("log level unspecified")
		return New("info", "file")
	}
}

func New(level, format string) *golog.Logger {
	logger := setupLogger(format)

	switch level {
	case "debug":
		golog.Info("logger level setup to debug")
		logger.SetLevel("debug")
	case "info":
		golog.Info("logger level setup to info")
		logger.SetLevel("info")
	case "warn":
		golog.Info("logger level setup to warn")
		logger.SetLevel("warn")
	case "error":
		golog.Info("logger level setup to error")
		logger.SetLevel("error")
	case "fatal":
		golog.Info("logger level setup to fatal")
		logger.SetLevel("fatal")
	default:
		golog.Info("logger level setup to info")
		logger.SetLevel("info")
	}

	return logger
}

func setupLogger(format string) *golog.Logger {
	switch format {
	case "json":
		golog.Info("setting json format for logger")
		moneybsLogger := golog.New()
		moneybsLogger.SetFormat("json", "    ")
		return moneybsLogger
	case "file":
		logFilePath := "./logs"

		err := os.MkdirAll(logFilePath, os.ModePerm)
		if err != nil {
			golog.Fatal("failed to create the log dir: ", err)
		}

		logFile := logFilePath + "/" + "moneybs.log"

		vlf, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o666)
		if err != nil {
			golog.Fatal("failed to create the log file due: ", err)
		}

		golog.Info("log filed created ", vlf.Name())

		golog.Info("setting file format for logger")
		moneybsLogger := golog.New()
		moneybsLogger.SetOutput(&lumberjack.Logger{
			Filename:   logFile,
			MaxAge:     logFileMaxAge,
			MaxSize:    logFileMaxSize,
			MaxBackups: logFileMaxBackups,
			Compress:   true,
		})
		moneybsLogger.SetFormat("json", "    ")

		return moneybsLogger
	case "stdout":
		moneybsLogger := golog.New()
		moneybsLogger.SetOutput(os.Stdout)

		return moneybsLogger
	default:
		moneybsLogger := golog.New()

		return moneybsLogger
	}
}
