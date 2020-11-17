package cmd

const(
	maxArgsLen = 0xfff
)


func PrepareArgs(origArgs []string) []string {
	args := make([]string, 0, len(origArgs))

	for i := range origArgs{
		args = append(args, origArgs[i])
	}
	
	return args
}

func init(){
	
}


