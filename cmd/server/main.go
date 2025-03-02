package main

import (
	"context"
	"os"

	"github.com/cloudsftp/botificator/pkg/server"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	logLevelStr := os.Getenv("LOG_LEVEL")
	if logLevelStr == "" {
		logLevelStr = "trace"
	}

	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		logrus.Fatalf("Invalid log level: %s", logLevelStr)
	}

	logrus.SetLevel(logLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	server, err := server.New(ctx)
	if err != nil {
		logrus.Fatalf("error in setup: %s", err)
	}
	defer server.Close()

	err = server.Run(ctx)
	if err != nil {
		logrus.Fatalf("error while running: %s", err)
	}
}
