all: proot-go

build: proot-go

proot-go: libproot
	go build -o proot-go

libproot:
	@make -C src libproot.a

clean:
	@make -C src clean
	rm -rf proot-go

rebuild: clean build