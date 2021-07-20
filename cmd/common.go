package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	PROOT_NAME = "proot"
)

var (
	logger = logrus.New()
)

func GoMain(args []string) {
	if v, ok := os.LookupEnv("PROOT_VERBOSE"); ok && (v == "1" || strings.ToLower(v) == "true") {
		logger.Level = logrus.DebugLevel
	}

	if len(args) == 0 {
		logger.Fatalln("no args passed")
	}

	binDir, binName := GetExecutableNameAndPath(args[0])
	goArgs, cArgs, cmdArgs := SplitArgs(binName, args[1:])

	for _, arg := range goArgs {
		if arg == "--go-verbose" {
			logger.Level = logrus.DebugLevel
		}
	}

	logger.Debugln("shell passed args: ", args)
	logger.Debugln("proot-go args: ", goArgs)
	logger.Debugln("proot args: ", cArgs)
	logger.Debugln("command args: ", cmdArgs)

	config := NewProotConfig()

	for _, arg := range goArgs {
		if arg == "--go-help" {
			logger.Debugln("dropping extra flags: ", cArgs, cmdArgs)

			config.Usage()

			cArgs = []string{"--help"}
			cmdArgs = []string{}
		}
	}

	if binName != PROOT_NAME {
		goArgs = append(goArgs, fmt.Sprintf("--go-config=%s", binDir))
		cmdArgs = append([]string{binName}, cmdArgs...)
	}

	if len(cmdArgs)>0{
		if err := config.Load(goArgs); err == nil {
			if cArgs, err := config.PrepareArgs(cArgs); err == nil {
				logger.Debugln("result args passed to c-proot: ", cArgs)
			}
		}else{
			logger.Warningln("error", err)
		}
	}

	cArgs = append([]string{PROOT_NAME}, cArgs...)

	cMain(append(cArgs, cmdArgs...))
}
