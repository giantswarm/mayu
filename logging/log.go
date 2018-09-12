package logging

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type MicrologgerWrapper struct {
	logger micrologger.Logger
}

func NewMicrologgerWrapper(logger micrologger.Logger) MicrologgerWrapper {
	return MicrologgerWrapper{
		logger: logger,
	}
}

func (l MicrologgerWrapper) Write(p []byte) (int, error) {
	err := l.logger.Log("level", "info", "type", "http log", "message", string(p))
	if err != nil {
		return 0, microerror.Mask(err)
	}
	return len(p), nil
}
