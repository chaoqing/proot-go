package main

// #cgo LDFLAGS: -L./src -lproot -ltalloc -larchive
//
// #include <stdlib.h>
// static void* alloc_string_slice(int len){
// return malloc(sizeof(char*)*len);
// }
//
// int proot_main(int argc, char *argv[]);
import "C"

import (
	"os"
	"proot_go/cmd"
	"unsafe"

	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

const (
	maxArgsLen = 0xfff
)

var (
	viperConfig = viper.New()
)

func cMain(args []string) {
	argc := C.int(len(args))

	log.Debugf("Got %v args: %v\n", argc, args)

	argv := (*[maxArgsLen]*C.char)(C.alloc_string_slice(argc))
	defer C.free(unsafe.Pointer(argv))

	for i, arg := range args {
		argv[i] = C.CString(arg)
		defer C.free(unsafe.Pointer(argv[i]))
	}

	C.proot_main(argc, (**C.char)(unsafe.Pointer(argv)))
}

func main() {
	config := &cmd.ProotConfig{}
	config.Load(viperConfig)

	args := cmd.PrepareArgs(os.Args, config)
	log.Info(args)

	cMain(args)
}

func init() {
	viperConfig.SetConfigName("proot")
	viperConfig.SetConfigType("yaml")

	viperConfig.AddConfigPath(".")

	cmd.ProotConfig{}.Register(viperConfig)

	if err := viperConfig.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(err)
		}
	}
}
