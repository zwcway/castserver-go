BIN_FILE=castserver
PWD=$(realpath .)

all: check front castserver

clean:
	rm -rf ${PWD}/web/public/*
	@go clean

build: castserver

castserver:
	@go build -o "${BIN_FILE}" .

check:
	@go fmt .
	@go vet .

front:
	cd ${PWD}/web/front && npm i && npm run build
