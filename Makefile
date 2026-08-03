BIN_FILE=castserver
PWD=$(realpath .)
NPM=npm

ifeq ($(OS), Windows_NT)
	NPM=npm.cmd
	BIN_FILE=castserver.exe
endif

all: check front castserver

clean:
	rm -rf ${PWD}/modules/web/public/*
	@go clean

build: castserver

castserver:
	@go build -o "${BIN_FILE}" .

check:
	@go fmt .
	@go vet .

front:
	cd ${PWD}/modules/web/front &&  ${NPM} i &&  ${NPM} run build
