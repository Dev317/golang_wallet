package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	HTTPServerAddress string `mapstructure:"http_server_address"`
	DBUser            string `mapstructure:"DB_USER"`
	DBPassword        string `mapstructure:"DB_PASSWORD"`
	DBName            string `mapstructure:"DB_NAME"`
	DBSSLMode         string `mapstructure:"DB_SSL_MODE"`
	DBHost            string `mapstructure:"DB_HOST"`
	DBPort            string `mapstructure:"DB_PORT"`
}

type QueueConfig struct {
	URI string `mapstructure:"URI"`
}

func LoadQueueConfig(path string) (config QueueConfig, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("queue")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}

func LoadServerConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("scanner")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
