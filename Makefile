all: proot-go

build: proot-go


GOBUILD = go build
GOENVS = env CGO_ENABLED=1
ifdef PI
GOENVS += env GOOS=linux GOARCH=arm GOARM=7
endif
ifdef Linux
GOENVS += env GOOS=linux GOARCH=amd64
endif

proot-go: libproot
	$(GOENVS) $(GOBUILD) -o proot-go

libproot: src/GNUmakefile
	@make -C src libproot.a

clean:
	@make -C src clean
	rm -rf proot-go

unit-test:
	go test -timeout 30s -v proot_go/cmd -run Test_prootFlagSet

rebuild: clean build