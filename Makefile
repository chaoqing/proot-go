all: proot-go

build: proot-go

proot-go: libproot
	go build -o proot-go

libproot: src/GNUmakefile
	@make -C src libproot.a

clean:
	@make -C src clean
	rm -rf proot-go

rebuild: clean build