package cmd

import (
	"github.com/spf13/viper"
)

const (
	maxArgsLen = 0xfff
)

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

func PrepareArgs(origArgs []string, config *ProotConfig) []string {
	args := make([]string, 0, len(origArgs))

	for i := range origArgs {
		args = append(args, origArgs[i])
	}

	return args
}

func init() {

}
