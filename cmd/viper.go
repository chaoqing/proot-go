package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	flag "github.com/spf13/pflag"
)

type ViperOption struct {
	OptionName   string
	EnvName      string
	FlagSuffix   string
	BindFlagName string
	Default      interface{}
	HelpMsg      string
}
type ViperOptions []*ViperOption

func (opt ViperOption) Register(conf *viper.Viper, flags *flag.FlagSet) {
	if conf != nil && opt.OptionName != "" {
		conf.SetDefault(opt.OptionName, opt.Default)
		if opt.EnvName != "" {
			_ = conf.BindEnv(opt.OptionName, opt.EnvName)
		}
	}

	if flags != nil && opt.FlagSuffix != "" {
		flagName := GO_ARGS_PREFIX[2:] + opt.FlagSuffix
		switch opt.Default.(type) {
		case []string:
			flags.StringArray(flagName, opt.Default.([]string), opt.HelpMsg)
		case string:
			flags.String(flagName, opt.Default.(string), opt.HelpMsg)
		case bool:
			flags.Bool(flagName, opt.Default.(bool), opt.HelpMsg)
		default:
			logger.Warningf("not recognized type on %v", opt)
		}

		if conf != nil {
			_ = conf.BindPFlag(opt.OptionName, flags.Lookup(opt.FlagSuffix))
		}
	}
}

var (
	prootGoOptions = ViperOptions{
		&ViperOption{
			OptionName:   "",
			EnvName:      "",
			FlagSuffix:   "help",
			BindFlagName: "",
			Default:      false,
			HelpMsg:      "Show proot-go help and exit",
		},
		&ViperOption{
			OptionName:   "",
			EnvName:      "",
			FlagSuffix:   "config",
			BindFlagName: "",
			Default:      "",
			HelpMsg:      "proot-go configuration path",
		},
		&ViperOption{
			OptionName:   "Verbose",
			EnvName:      "VERBOSE",
			FlagSuffix:   "verbose",
			BindFlagName: "",
			Default:      false,
			HelpMsg:      "Enabling proot-go logging",
		},
		&ViperOption{
			OptionName:   "RootPath",
			EnvName:      "ROOT_PATH",
			FlagSuffix:   "root",
			BindFlagName: "-r",
			Default:      "",
			HelpMsg:      "Same as proot '-r *path*' and use '--go-root $$[R|S]:*path*' to bind with '-[R|S]' instead of '-r'",
		},
		&ViperOption{
			OptionName:   "Bind",
			EnvName:      "BIND",
			FlagSuffix:   "bind",
			BindFlagName: "-b",
			Default:      []string{},
			HelpMsg:      "Same as proot '-b *path*'",
		},
		&ViperOption{
			OptionName:   "WorkDirectory",
			EnvName:      "CWD",
			FlagSuffix:   "cwd",
			BindFlagName: "w",
			Default:      "$$CWD",
			HelpMsg:      "Same as proot '-w *path*'",
		},
		&ViperOption{
			OptionName:   "DirectoryMap",
			EnvName:      "DIR_MAP",
			FlagSuffix:   "dir-map",
			BindFlagName: "",
			Default:      []string{"$$BIND"},
			HelpMsg:      "The direcotry map apply to *PATH like Envs and default cwd, use '$$BIND' to include all bind options(which include '-r' bind)",
		},
		&ViperOption{
			OptionName:   "Env",
			EnvName:      "ENV",
			FlagSuffix:   "env",
			BindFlagName: "",
			Default:      []string{},
			HelpMsg:      "Start proot client process with prefix 'env ENV_NAME=ENV_VALUE' if any environment set",
		},
		&ViperOption{
			OptionName:   "RawOption",
			EnvName:      "RAW_OPTION",
			FlagSuffix:   "raw",
			BindFlagName: "",
			Default:      []string{},
			HelpMsg:      "Raw options which will pass to proot like '-R *path* -S *path*'",
		},
	}
)

type ProotConfig struct {
	RootPath       string
	RootBindOption string

	Bind []string

	WorkDirectory string

	DirectoryMap []string

	Env []string

	RawOption []string

	viperConfig *viper.Viper
	flagConfig  *flag.FlagSet
}

func (config *ProotConfig) Register() *ProotConfig {
	if config.flagConfig == nil {
		config.flagConfig = flag.NewFlagSet(PROOT_NAME, flag.ContinueOnError)
	}

	if config.viperConfig == nil {
		config.viperConfig = viper.New()
		config.viperConfig.SetEnvPrefix(PROOT_NAME)
	}

	// for _, opt := range prootGoOptions {
	// 	opt.Register(config.viperConfig, config.flagConfig)
	// }

	return config
}

func (config ProotConfig) Usage() {
	usage := `PRoot-go has a go command wrapper on original C version of proot to provide smart interface.

Usage:
	proot-go [--go-option]... [--proot-option]... [command]
	busybox [command-option]

Flags:
%s

%s
Raw *proot* helper:

	`

	exampleContent := ""

	if file, err := ioutil.TempFile("", "proot_*.yaml"); err == nil {
		examplePath := file.Name()
		logger.Debugln("using temporay file ", examplePath)
		file.Close()
		defer os.Remove(examplePath)

		v := config.viperConfig
		v.SetConfigType("yaml")
		_ = v.ReadConfig(bytes.NewReader([]byte("RootPath: /")))
		if err := v.WriteConfigAs(examplePath); err == nil {
			if content, err := ioutil.ReadFile(examplePath); err == nil {
				exampleContent = fmt.Sprintf(`Example YAML configura file:
----------------------------
%s
----------------------------
`, string(content))
			}
		}
	}

	fmt.Printf(usage, config.flagConfig.FlagUsages(), exampleContent)
}

func (config *ProotConfig) Load(args []string) error {
	if err := config.flagConfig.Parse(args); err != nil {
		logger.Fatalln("proot go command args error: ", err)
	}

	config.viperConfig.SetConfigName(PROOT_NAME)
	config.viperConfig.SetConfigType("yaml")

	configDir := "."
	configPath := ""
	if path, err := config.flagConfig.GetString("go-config"); err == nil && path != "" {
		configPath = path
	} else if path, ok := os.LookupEnv("PROOT_CONFIG"); ok && path != "" {
		configPath = path
	}
	if configPath != "" {
		if info, err := os.Stat(configPath); err == nil && info.IsDir() {
			config.viperConfig.AddConfigPath(configPath)
			configPath = ""
			configDir = ""
		} else if err == nil {
			if content, err := ioutil.ReadFile(configPath); err == nil {
				if err := config.viperConfig.ReadConfig(bytes.NewReader(content)); err == nil {
					configDir = "" // to mark the reading is done
				} else {
					logger.Fatalf("error when loading config file <%s>: %v", configPath, err)
				}
			} else {
				logger.Fatalf("error when reading config file <%s>: %v", configPath, err)
			}
		} else {
			logger.Fatalf("can not find config file path or directory <%s>: %v", configPath, err)
		}
	}

	if configDir != "" {
		for {
			if info, err := os.Stat(configPath); err == nil && info.IsDir() {
				config.viperConfig.AddConfigPath(configPath)

				configPath = filepath.Join("..", configPath)
			} else {
				break
			}
		}
	}

	if configPath == "" {
		if err := config.viperConfig.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return fmt.Errorf("proot-go config file not found")
			} else {
				logger.Fatalln("proot-go config load error: ", err)
			}
		}
	}

	if err := config.viperConfig.Unmarshal(config); err != nil {
		logger.Fatalln("proot-go config load error: ", err)
	}

	return nil
}

func (config *ProotConfig) PrepareArgs(args []string) ([]string, error) {
	return args, nil
}

func init() {
}
