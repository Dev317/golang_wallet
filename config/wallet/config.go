package config

import "github.com/spf13/viper"

type Config struct {
	HTTPServerAddress string `mapstructure:"HTTP_SERVER_ADDRESS"`
	DBUser            string `mapstructure:"DB_USER"`
	DBPassword        string `mapstructure:"DB_PASSWORD"`
	DBName            string `mapstructure:"DB_NAME"`
	DBSSLMode         string `mapstructure:"DB_SSL_MODE"`
	DBHost            string `mapstructure:"DB_HOST"`
	DBPort            string `mapstructure:"DB_PORT"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("wallet")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
