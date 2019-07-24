package log

import (
	"github.com/sirupsen/logrus"
)

func init() {
	formatter := new(logrus.TextFormatter)
	formatter.TimestampFormat = "2006-01-02 15:04:05"
	formatter.FullTimestamp = true
	logrus.SetFormatter(formatter)
	logrus.SetLevel(logrus.InfoLevel)
}
