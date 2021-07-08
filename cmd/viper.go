package cmd

import (
	"github.com/spf13/viper"
)

var (
	viperConfig = viper.New()
)

const VIPER_YAML_EXAMPLE = `
RootDir: /root

HostHome: /home
GuestHome: /root

Env:
  - PATH=$PATH

ExtraOptions:
`

type ProotConfig struct {
	RootDir string

	HostHome  string
	GuestHome string

	Env []string

	ExtraOptions []string
}

func (config ProotConfig) Register(v *viper.Viper) {
	v.SetEnvPrefix("proot")

}

func (config *ProotConfig) Load(v *viper.Viper) *ProotConfig {
	v.Unmarshal(config)

	return config
}

func init() {
	viperConfig.SetConfigName("proot")
	viperConfig.SetConfigType("yaml")

	viperConfig.AddConfigPath(".")

	ProotConfig{}.Register(viperConfig)

	if err := viperConfig.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(err)
		}
	}
}
