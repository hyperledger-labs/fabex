package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Mongo struct {
	Host       string
	Port       int
	Dbuser     string
	Dbsecret   string
	Dbname     string
	Collection string
}

type GRPCServer struct {
	Host string
	Port string
}

type Fabric struct {
	User              string
	Secret            string
	Org               string
	Channels          []string
	ConnectionProfile string
}

type UI struct {
	Host string
	Port string
}

type Config struct {
	Mongo      `mapstructure:"mongo"`
	Fabric     `mapstructure:"fabric"`
	GRPCServer `mapstructure:"grpc"`
	UI         `mapstructure:"ui"`
}

type BootConfig struct {
	Enrolluser bool   `env:"ENROLL" envDefault:"false"`
	Config     string `env:"CONFIG"`
	Database   string `env:"DB" envDefault:"mongo"`
	UI         bool   `env:"UI" envDefault:"true"`
	LogLevel   string `env:"LOG" envDefault:"info"`
}

func GetBootConfig() (*BootConfig, error) {
	var bootConf BootConfig
	if err := env.Parse(&bootConf); err != nil {
		return nil, errors.WithStack(err)
	}
	return &bootConf, nil
}

func GetMainConfig(bootConfig *BootConfig) (*Config, error) {
	viper.SetConfigFile(bootConfig.Config)

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.WithStack(err)
	}

	mainConf := &Config{}
	if err := viper.UnmarshalExact(mainConf); err != nil {
		return nil, errors.WithStack(errors.Wrapf(err, "unable to unmarshal main config %s", bootConfig.Config))
	}

	return mainConf, nil
}
