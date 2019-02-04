package log

import "github.com/sirupsen/logrus"

func ForSubsystem(subsystem string) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"subsystem": subsystem,
	})
}
