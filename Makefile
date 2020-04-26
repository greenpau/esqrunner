.PHONY: test clean qtest
APP=esqrunner
APP_VERSION:=$(shell cat VERSION | head -1)
GIT_COMMIT:=$(shell git describe --dirty --always)
GIT_BRANCH:=$(shell git rev-parse --abbrev-ref HEAD -- | head -1)
BUILD_USER:=$(shell whoami)
BUILD_DATE:=$(shell date +"%Y-%m-%d")
BINARY:=esqrunner
VERBOSE:=-v

all:
	@echo "Version: $(APP_VERSION), Branch: $(GIT_BRANCH), Revision: $(GIT_COMMIT)"
	@echo "Build on $(BUILD_DATE) by $(BUILD_USER)"
	@mkdir -p bin/
	@rm -rf ./bin/*
	@CGO_ENABLED=0 go build -o ./bin/$(BINARY) $(VERBOSE) \
		-ldflags="-w -s \
		-X main.appVersion=$(APP_VERSION) \
		-X main.gitBranch=$(GIT_BRANCH) \
		-X main.gitCommit=$(GIT_COMMIT) \
		-X main.buildUser=$(BUILD_USER) \
		-X main.buildDate=$(BUILD_DATE)" \
		-gcflags="all=-trimpath=$(GOPATH)/src" \
		-asmflags="all=-trimpath $(GOPATH)/src" \
		*.go
	@echo "Done!"

test: all
	@golint -set_exit_status *.go
	@echo "PASS: Golang Lint test"
	@for f in `find ./assets -name *.y*ml`; do yamllint $$f; done
	@echo "PASS: YAML Lint test"
	@go test -v *.go
	@echo "PASS: core tests"
	@./bin/$(BINARY) --config ./assets/conf/default.yaml --validate --log-level debug
	@./bin/$(BINARY) --version
	@echo "PASS: configuration validation"
	@echo "OK: all tests passed!"

clean:
	@rm -rf bin/ build/ pkg-build/
	@echo "OK: clean up completed"

qtest:
	@#./bin/$(BINARY) -config ./assets/conf/$(BINARY).yaml -log-level debug -dry-run
	@./bin/$(BINARY) -config ./assets/conf/$(BINARY).yaml -log-level debug
