package logging

import (
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

// SetLevel :
func SetLevel(log *logrus.Logger, level string) error {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(lvl)
	return nil
}


// GetLogger :
func GetLogger() *logrus.Logger {
	return logger
}

func init() {
	logger = logrus.StandardLogger()
}
