package log

import "github.com/sirupsen/logrus"

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func ForSubsystem(subsystem string) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"subsystem": subsystem,
	})
}

func WithError(logger *logrus.Entry, err error) *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"err": err,
	})
}