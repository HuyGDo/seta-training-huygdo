package logger

import (
	"io"
	"os"
	"path"

	"github.com/rs/zerolog"
)

func New() *zerolog.Logger {
	// Create the logs directory if it doesn't exist
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.Mkdir("logs", 0755)
	}

	logFile, err := os.OpenFile(
		path.Join("logs", "app.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	if err != nil {
		panic(err)
	}

	// Use MultiLevelWriter to log to both console and file
	writer := io.MultiWriter(os.Stdout, logFile)
	log := zerolog.New(writer).With().Timestamp().Logger()

	return &log
}
