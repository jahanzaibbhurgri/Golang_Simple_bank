package utils

import "github.com/spf13/viper"

//config stores all configuration of the application
// The values are read by viper from a config file or environment variable
type Config struct {
    DBDriver      string `mapstructure:"DB_DRIVER"` //this would be referencing to the env
    DBSource      string `mapstructure:"DB_SOURCE"`
    ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}

//LoadConfig read configuration from file or env variable//
func LoadConfig(path string) (config Config, err error) {
    viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env") //json,xml
	
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
	  return 
	}
	err = viper.Unmarshal(&config)
	return 
}