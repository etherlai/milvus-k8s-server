package log

import (
	"github.com/sirupsen/logrus"
)

var (
	std    = logrus.New()
	Logger *logrus.Logger
)

func Init() {
	std.SetReportCaller(true)
	Logger = std
}
