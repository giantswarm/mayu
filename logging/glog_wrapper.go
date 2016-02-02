package logging

import (
	"github.com/golang/glog"
)

type GlogWrapper struct {
	verbosityLevel glog.Level
}

func NewGlogWrapper(level glog.Level) GlogWrapper {
	return GlogWrapper{
		verbosityLevel: level,
	}
}

func (gw GlogWrapper) Write(p []byte) (int, error) {
	glog.V(gw.verbosityLevel).Info(string(p))

	return len(p), nil
}
