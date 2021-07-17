package cmd

// #cgo LDFLAGS: -L../src -lproot -ltalloc -larchive
//
// #include <stdlib.h>
// static void* alloc_string_slice(int len){
// return malloc(sizeof(char*)*len);
// }
//
// int proot_main(int argc, char *argv[]);
import "C"

import (
	"unsafe"
)

const (
	maxArgsLen = 0xfff
)


func cMain(args []string) {
	argc := C.int(len(args))

	logger.Debugf("Got %v args: %v\n", argc, args)

	argv := (*[maxArgsLen]*C.char)(C.alloc_string_slice(argc))
	defer C.free(unsafe.Pointer(argv))

	for i, arg := range args {
		argv[i] = C.CString(arg)
		defer C.free(unsafe.Pointer(argv[i]))
	}

	C.proot_main(argc, (**C.char)(unsafe.Pointer(argv)))
}