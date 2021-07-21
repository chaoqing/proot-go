package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/viper"
	"github.com/yookoala/realpath"

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

func (opt ViperOption) Register(conf *viper.Viper, flags *flag.FlagSet) {
	if conf != nil && opt.OptionName != "" {
		conf.SetDefault(opt.OptionName, opt.Default)
		if opt.EnvName != "" {
			_ = conf.BindEnv(opt.OptionName, opt.EnvName)
		}
	}

	flagName := ""

	if flags != nil && opt.FlagSuffix != "" {
		flagName = GO_ARGS_PREFIX[2:] + opt.FlagSuffix
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

		if conf != nil && opt.OptionName != "" && flagName != ""{
			_ = conf.BindPFlag(opt.OptionName, flags.Lookup(flagName))
		}
	}
}

type ViperOptions []*ViperOption

func (opts ViperOptions) Register(conf *viper.Viper, flags *flag.FlagSet){
	for _, opt := range opts {
		opt.Register(conf, flags)
	}
}

var (
	ProotGoOptions = ViperOptions{
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
			Default:      "$$R/",
			HelpMsg:      "Same as proot '-r *path*' and use '--go-root $$[R|S]:*path*' to bind with '-[R|S]' instead of '-r'",
		},
		&ViperOption{
			OptionName:   "Bind",
			EnvName:      "",
			FlagSuffix:   "bind",
			BindFlagName: "-b",
			Default:      []string{"$$ENV:/usr/bin/env"},
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
			EnvName:      "",
			FlagSuffix:   "map",
			BindFlagName: "",
			Default:      []string{"$$BIND"},
			HelpMsg:      "The direcotry map apply to *PATH like Envs and default cwd, use '$$BIND' to include all bind options(which include '-r' bind)",
		},
		&ViperOption{
			OptionName:   "Env",
			EnvName:      "",
			FlagSuffix:   "env",
			BindFlagName: "",
			Default:      []string{"PATH","LANG"},
			HelpMsg:      "Start proot client process with prefix 'env ENV_NAME=ENV_VALUE' if any environment set",
		},
		&ViperOption{
			OptionName:   "RawOption",
			EnvName:      "",
			FlagSuffix:   "raw",
			BindFlagName: "",
			Default:      []string{"-0"},
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
	FlagConfig  *flag.FlagSet
}

func NewProotConfig(goArgs []string) *ProotConfig {
	config := &ProotConfig{
		viperConfig: viper.New(),
		FlagConfig: flag.NewFlagSet(PROOT_NAME, flag.ContinueOnError),
	}
	config.viperConfig.SetEnvPrefix(PROOT_NAME)

	ProotGoOptions.Register(config.viperConfig, config.FlagConfig)

	if err := config.FlagConfig.Parse(goArgs); err != nil {
		logger.Fatalln("proot go command args error: ", err)
	}

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
		logger.Traceln("using temporay file ", examplePath)
		file.Close()
		defer os.Remove(examplePath)

		v := config.viperConfig
		v.SetConfigType("yaml")
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

	fmt.Printf(usage, config.FlagConfig.FlagUsages(), exampleContent)
}

func (config *ProotConfig) Load() error {
	v := config.viperConfig

	v.SetConfigName(fmt.Sprintf(".%s", PROOT_NAME))
	v.SetConfigType("yaml")

	configDir := "."
	configPath := ""
	if path, err := config.FlagConfig.GetString("go-config"); err == nil && path != "" {
		configPath = path
	} else if path, ok := os.LookupEnv("PROOT_CONFIG"); ok && path != "" {
		configPath = path
	}
	if configPath != "" {
		if info, err := os.Stat(configPath); err == nil && info.IsDir() {
			v.AddConfigPath(configPath)
			configPath = ""
			configDir = ""
		} else if err == nil {
			logger.Debugln("configure file path: ", configPath)
			v.SetConfigFile(configPath)

			if err := v.ReadInConfig(); err == nil {
				configDir = "" // to mark the reading is done
			} else {
				logger.Fatalf("error when reading config file <%s>: %v", configPath, err)
			}
		} else {
			logger.Fatalf("can not find config file path or directory <%s>: %v", configPath, err)
		}
	}

	if configDir != "" {
		for {
			if info, err := os.Stat(configDir); err == nil && info.IsDir() {
				v.AddConfigPath(configDir)

				configDir = filepath.Join("..", configDir)
			} else {
				break
			}
		}
	}

	if configPath == "" {
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				return fmt.Errorf("proot-go config file not found")
			}
			logger.Fatalln("proot-go config load error: ", err)
		}else{
			logger.Debugln("found configure file ", v.ConfigFileUsed())
		}
	}

	if err := v.Unmarshal(config); err != nil {
		logger.Fatalln("proot-go config load error: ", err)
	}

	return nil
}

const ENV = "env"
func searchEnv() string{
	if envPath, err := exec.LookPath(ENV); err==nil{
		if path, err := realpath.Realpath(envPath); err==nil{
			return path
		}
	}

	return ""
}

type  dirMapRelation struct{
	hostDir string
	clientDir string
}

type  dirMapRelations []dirMapRelation
func (a dirMapRelations) Len() int {
	return len(a)
}
  
func (a dirMapRelations) Less(i, j int) bool {
	return len(a[i].hostDir) < len(a[j].hostDir)
}

func (a dirMapRelations) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func CleanPathMap(host2client map[string]string) func(string) string{
	mapRelation := make(dirMapRelations, 0, len(host2client))
	for k, v := range host2client{
		mapRelation = append(mapRelation, dirMapRelation{filepath.Clean(k), filepath.Clean(v)})
	}
	sort.Sort(mapRelation)

	return func(p string) string{
		for _, r := range(mapRelation){
			if strings.HasPrefix(p, r.hostDir){
				return strings.Replace(p, r.hostDir, r.clientDir, 1)
			}
		}
		return p
	}
}

func (config *ProotConfig) PrepareArgs(args []string) ([]string, error) {
	logger.Debugln("before process: ", config)

	if strings.HasPrefix(config.RootPath, "$$R"){
		config.RootPath = config.RootPath[len("$$R"):]
		config.RootBindOption = "-R"
	}else if strings.HasPrefix(config.RootPath, "$$S"){
		config.RootPath = config.RootPath[len("$$S"):]
		config.RootBindOption = "-S"
	}else{
		config.RootBindOption = "-r"
	}

	bindMap := make(map[string]string)
	for i, s := range config.Bind{
		if  strings.HasPrefix(s, "$$ENV"){
			if envPath := searchEnv(); envPath!=""{
				config.Bind[i] = strings.Replace(s, "$$ENV", envPath, 1)
			}else{
				logger.Fatalln("can not find executable env")
			}
		}

		if m := filepath.SplitList(config.Bind[i]); len(m)==2{
			bindMap[m[0]] = m[1]
		}
	}

	directoryMap := make(map[string]string)
	for _, s := range config.DirectoryMap{
		if s=="$$BIND"{
			for k,v := range bindMap{
				directoryMap[k] = v
			}
		}else{
			kv := strings.SplitN(s, "=", 2)

			if  len(kv)!=1{
				logger.Fatalln("directory map relation must in form of 'HostDir=ClientDir' but got: ", s)
			}
			
			directoryMap[kv[0]] = kv[1]
		}
	}

	directoryMapFunc := CleanPathMap(directoryMap)

	if config.WorkDirectory == "$$CWD"{
		if cwd, err := os.Getwd(); err==nil{
			config.WorkDirectory = cwd
		}
	}
	config.WorkDirectory = directoryMapFunc(config.WorkDirectory)

	for i, s := range config.Env{
		kv := strings.SplitN(s, "=", 2)

		if  len(kv)==1{
			kv = append(kv, os.Getenv(kv[0]))
		}

		if strings.HasSuffix(kv[0], "PATH"){
			paths := filepath.SplitList(kv[1])
			for j, p := range paths{
				paths[j] = directoryMapFunc(p)
			}
			kv[1] = strings.Join(paths, string(filepath.ListSeparator))
		}

		config.Env[i] = strings.Join(kv, "=")
	}

	logger.Debugln("after process: ", config)

	newArgs := make([]string, 0, len(args)*2)
	newArgs = append(newArgs, config.RawOption...)
	newArgs = append(newArgs, config.RootBindOption, config.RootPath)

	for _, v := range config.Bind{
		newArgs = append(newArgs, "-b", v)
	}

	newArgs = append(newArgs, "-w", config.WorkDirectory)

	newArgs = append(newArgs, args...)

	if len(config.Env)>0{
		newArgs = append(newArgs, ENV)
		newArgs = append(newArgs, config.Env...)
	}

	return newArgs, nil
}

func init() {
}
