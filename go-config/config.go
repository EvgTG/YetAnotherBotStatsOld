package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	viper *viper.Viper
}

func New() *Config {
	viper := viper.New()
	viper.AutomaticEnv()

	if configName := viper.GetString("CONFIG_NAME"); configName != "" {
		godotenv.Load(configName)
	}

	return &Config{viper: viper}
}

func (c *Config) GetString(key string) string {
	return c.viper.GetString(key)
}

func (c *Config) IsSet(key string) bool {
	return c.viper.IsSet(key)
}

func (c *Config) GetInt(key string) int {
	return c.viper.GetInt(key)
}

/*
Get(key string) : interface{}
GetBool(key string) : bool
GetFloat64(key string) : float64
GetInt(key string) : int
GetIntSlice(key string) : []int
GetString(key string) : string
GetStringMap(key string) : map[string]interface{}
GetStringMapString(key string) : map[string]string
GetStringSlice(key string) : []string
GetTime(key string) : time.Time
GetDuration(key string) : time.Duration
IsSet(key string) : bool
*/
