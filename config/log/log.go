package log

import (
	"github.com/hyperledger-labs/fabex/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func GetLogger(config *models.BootConfig) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(config.LogLevel)
	if err != nil {
		return nil, errors.WithStack(errors.Wrap(err, "failed to parse log level"))
	}
	l := zap.Config{
		Level:             lvl,
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "console"}
	return l.Build()
}
