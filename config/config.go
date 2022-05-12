package config

import (
	"github.com/hyperledger-labs/fabex/models"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func GetBootConfig() (*models.BootConfig, error) {
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.WithStack(errors.Wrap(err, "unable to read boot config"))
	}

	var bootConf models.BootConfig
	err := viper.Unmarshal(&bootConf)
	if err != nil {
		return nil, errors.WithStack(errors.Wrap(err, "unable to unmarshal boot config"))
	}
	return &bootConf, err
}

func GetMainConfig(bootConfig *models.BootConfig) (*models.Config, error) {
	viper.SetConfigFile(bootConfig.Confpath)
	viper.SetConfigType("yaml")

	var mainConf models.Config
	if err := viper.Unmarshal(&mainConf); err != nil {
		return nil, errors.WithStack(errors.Wrapf(err, "unable to unmarshal main config %s", bootConfig.Confpath))
	}
	return &mainConf, nil
}
