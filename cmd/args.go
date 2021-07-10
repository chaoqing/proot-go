package cmd

import (
	"path/filepath"
	"regexp"
	"strings"

	flag "github.com/spf13/pflag"
)

const GO_ARGS_PREFIX = "--go-"

func prootFlagSet() *flag.FlagSet {
	// `./proot-go --help |grep '^ *-'`
	const orignalArgsHelp = `
	-r *path*, --rootfs=*path*
	-b *path*, --bind=*path*, -m *path*, --mount=*path*
	-q *command*, --qemu=*command*
	-w *path*, --pwd=*path*, --cwd=*path*
	--kill-on-exit
	-v *value*, --verbose=*value*
	-V, --version, --about
	-h, --help, --usage
	-k *string*, --kernel-release=*string*
	-0, --root-id
	-i *string*, --change-id=*string*
	-p *string*, --port=*string*
	-n, --netcoop
	-l, --link2symlink
	-R *path*
	-S *path*
	`

	prootFlags := flag.NewFlagSet(PROOT_NAME, flag.ContinueOnError)

	boolLongFlags := make([]string, 0)
	boolShortFlags := make([]string, 0)

	stringLongFlags := make([]string, 0)
	stringShortFlags := make([]string, 0)

	flagRegex := regexp.MustCompile(`(-[\w]|--\w[\w-]{1,})([ =]\*[^\*]+\*)?(,|$)`)
	argsHelp := strings.ReplaceAll(strings.TrimSpace(orignalArgsHelp), "\n", ",")
	for _, arg := range flagRegex.FindAllStringSubmatch(argsHelp, -1){
		logger.Debugln(arg[1], arg[0])
		if len(arg[2])==0 {
			if len(arg[1])==2{
				boolShortFlags = append(boolShortFlags, arg[1][1:])
			}else{
				boolLongFlags = append(boolLongFlags, arg[1][2:])
			}
		}else{
			if len(arg[1])==2{
				stringShortFlags = append(stringShortFlags, arg[1][1:])
			}else{
				stringLongFlags = append(stringLongFlags, arg[1][2:])
			}
		}
	}

	nFlags := len(boolShortFlags)
	if nFlags > len(boolLongFlags){
		nFlags = len(boolLongFlags)
	}
	for i, arg := range boolShortFlags[:nFlags]{
		prootFlags.BoolP(boolLongFlags[i], arg, false, "")
	}
	for _, arg := range boolLongFlags[nFlags:]{
		prootFlags.Bool(arg, false, "")
	}
	for _, arg := range boolShortFlags[nFlags:]{
		prootFlags.BoolS(arg, arg, false, "")
	}

	nFlags = len(stringShortFlags)
	if nFlags > len(stringLongFlags){
		nFlags = len(stringLongFlags)
	}
	for i, arg := range stringShortFlags[:nFlags]{
		prootFlags.StringP(stringLongFlags[i], arg, "", "")
	}
	for _, arg := range stringLongFlags[nFlags:]{
		prootFlags.String(arg, "", "")
	}
	for _, arg := range stringShortFlags[nFlags:]{
		prootFlags.StringS(arg, arg, "", "")
	}

	return prootFlags
}

func splitArgs(binName string, args []string) (goArgs []string, cArgs []string, cmdArgs []string){
	goArgs = make([]string, 0, 5)
	cArgs = make([]string, 0, len(args))
	
	if binName != PROOT_NAME{
		cmdArgs = append([]string{binName}, args...)
		return goArgs, cArgs, cmdArgs
	}

	cFlagSet := prootFlagSet()
	for i, arg := range args{
		if strings.HasPrefix(strings.ToLower(arg), GO_ARGS_PREFIX){
			goArgs = append(goArgs, arg)
		}else{
			cArgs = append(cArgs, arg)
			if err := cFlagSet.Parse(cArgs); err==nil && cFlagSet.NArg()==1{
				cArgs = cArgs[:(len(cArgs)-1)]
				cmdArgs = args[i:]

				break
			}
		}
	}
	return goArgs, cArgs, cmdArgs
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
	if len(origArgs) == 0{
		logger.Fatalln("no args passed")
	}
	
	binName := GetExecutableName(origArgs[0])
	goArgs, cArgs, cmdArgs := splitArgs(binName, origArgs[1:])

	for _, arg := range goArgs{
		if arg == "--go-help"{
			logger.Infoln("Print --go-help here: ", goArgs, VIPER_YAML_EXAMPLE)
			logger.Debugln("dropping extra flags: ", cArgs, cmdArgs)

			return []string{PROOT_NAME, "--help"}
		}
	}
	
	cArgs = append([]string{PROOT_NAME}, cArgs...)
	return append(cArgs, cmdArgs...)
}

func init(){
	
}

