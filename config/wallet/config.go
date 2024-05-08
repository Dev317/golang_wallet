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

type ChainItemConfig struct {
	ChainID   string `mapstructure:"chain_id"`
	ChainName string `mapstructure:"chain_name"`
	RPCURL    string `mapstructure:"rpc_url"`
}

type EthereumConfig struct {
	ChainItemList []ChainItemConfig `yaml:"ChainItemList,mapstructure"`
}

type QueueConifg struct {
	URI string `mapstructure:"URI"`
}

func LoadQueueConfig(path string) (config QueueConifg, err error) {
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

func LoadEthereumConfig(path string) (config EthereumConfig, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("ethereum")
	viper.SetConfigType("yaml")

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}

func LoadServerConfig(path string) (config Config, err error) {
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
