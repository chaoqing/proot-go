all: proot-go

build: proot-go

proot-go: libproot
	go build -o proot-go

libproot: src/GNUmakefile
	@make -C src libproot.a

clean:
	@make -C src clean
	rm -rf proot-go

unit-test:
	go test -timeout 30s -v proot_go/cmd -run Test_prootFlagSet

rebuild: clean build