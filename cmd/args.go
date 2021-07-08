package cmd

import (
	"path/filepath"
	"regexp"
	"strings"	
)

func splitArgs(origArgs []string) (goArgs []string, cArgs []string){
	goArgs = make([]string, 0, len(origArgs))
	cArgs = make([]string, 0, len(origArgs))

	for _, arg := range origArgs{
		if strings.HasPrefix(arg, "--go-"){
			goArgs = append(goArgs, arg)
		}else{
			cArgs = append(cArgs, arg)
		}
	}

	return goArgs, cArgs
}

func GetExecutableName(path string) string {
	_, name := filepath.Split(path)

	if strings.HasPrefix(strings.ToLower(name), PROOT_NAME){
		return PROOT_NAME
	}else{
		return name
	}
}


func PrepareArgs(origArgs []string) []string {
	if len(origArgs)==0{
		logger.Panicln("zero args passed")
	}

	binPath := origArgs[0]
	binName := GetExecutableName(binPath)
	goArgs, cArgs := splitArgs(origArgs[1:])

	if binName == PROOT_NAME{
		hasHelp := false
		re := regexp.MustCompile("(-h|--help|--usage)")
		for _, arg := range origArgs[1:]{
			if ok := re.MatchString(strings.TrimPrefix(arg, "--go")); ok{
				hasHelp = true
				break
			}
		}
		if hasHelp{
			logger.Infoln("Print --go--help here: ", binPath, goArgs, VIPER_YAML_EXAMPLE)
			logger.Debugln("dropping c-args: ", cArgs)
			cArgs =  []string{"--help"}
		}
	}else{
		cArgs = append(cArgs, binName)
	}
	
	return append([]string{PROOT_NAME}, cArgs...)
}

func init(){
	
}

