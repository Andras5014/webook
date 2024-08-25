package startup

import (
	"github.com/Andras5014/webook/config"
	"github.com/spf13/viper"
)

func InitConfig() *config.Config {
	cfg := &config.Config{}
	err := viper.Unmarshal(cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}
