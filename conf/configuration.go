package conf

import (
	"fmt"
	"os"
	"sync"

	"github.com/DE-labtory/swim"

	"github.com/spf13/viper"
)

var confPath = os.Getenv("GOPATH") + "/src/github.com/DE-labtory/swim/conf/config.yaml"

type Configuration struct {
	SWIMConfig            swim.Config
	SuspicionConfig       swim.SuspicionConfig
	MessageEndpointConfig swim.MessageEndpointConfig
	Member                swim.Member
}

// config is instance of SWIM configuration
var config = &Configuration{}

var once = sync.Once{}

func SetConfigPath(abspath string) {
	confPath = abspath
}

func GetConfiguration() *Configuration {
	once.Do(func() {
		viper.SetConfigFile(confPath)
		if err := viper.ReadInConfig(); err != nil {
			panic(fmt.Sprintf("cannot read config from %s", confPath))
		}
		err := viper.Unmarshal(&config)
		if err != nil {
			panic("error in read config")
		}
	})

	return config
}
