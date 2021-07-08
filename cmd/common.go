package cmd

import (
	"github.com/sirupsen/logrus"
)

const(
	PROOT_NAME = "proot"
)

var(
	logger = logrus.New()
)

func detectAndSetLoggingLevel(origArgs []string){
	for _, arg := range origArgs{
		if arg == "--go--verbose"{
			logger.Level = logrus.TraceLevel
		}
	} 
}

func GoMain(args [] string) {
	config := &ProotConfig{}
	config.Load(viperConfig)

	detectAndSetLoggingLevel(args)

	logger.Debugln("shell passed args: ", args)
	cArgs := PrepareArgs(args)
	logger.Debugln("result args passed to c-proot: ", cArgs)
	cMain(cArgs)
}