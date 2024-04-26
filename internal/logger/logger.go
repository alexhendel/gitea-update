package logger

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func InitLogger() {
	Logger.Out = os.Stdout
	Logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "02.01.2006 15:04:05.000",
		FullTimestamp:   true,
	})
}

func LogWarn(msg string) {
	Logger.WithTime(time.Now()).Warn(msg)
}

func LogInfo(msg string) {
	Logger.WithTime(time.Now()).Info(msg)
}

func LogError(msg string) {
	Logger.WithTime(time.Now()).Error(msg)
}

func LogFatal(msg string) {
	Logger.WithTime(time.Now()).Fatal(msg)
}

func LogDebug(msg string) {
	Logger.WithTime(time.Now()).Debug(msg)
}
