ifndef $(GOPATH)
	GOPATH=$(shell go env GOPATH)
endif
SHELL = bash

%.o: %.mod

run: build
	@echo "starting service"
	./bin/runner

build: 
	@echo "building started"
	go build -o bin/runner cmd/main.go
	@echo "project builded"

bench:
	./sh/run_bench.sh bench_tests.list
