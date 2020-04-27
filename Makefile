.PHONY: test clean qtest linter covdir integration coverage
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
		cmd/$(APP)/*.go
	@echo "Done!"

linter:
	@echo "Running lint checks"
	@golint -set_exit_status *.go
	@golint -set_exit_status cmd/$(APP)/*.go
	@for f in `find ./assets -name *.y*ml`; do yamllint $$f; done
	@echo "PASS: lint checks"

covdir:
	@echo "Creating .coverage/ directory"
	@mkdir -p .coverage

integration: all
	@echo "Running integration tests"
	@./bin/$(BINARY) --config ./assets/conf/default.yaml --validate --log-level debug
	@./bin/$(BINARY) --version


test: covdir linter integration
	@go test -v -coverprofile=.coverage/coverage.out ./*.go

coverage:
	@go tool cover -html=.coverage/coverage.out -o .coverage/coverage.html
	@go test -covermode=count -coverprofile=.coverage/coverage.out ./*.go
	@go tool cover -func=.coverage/coverage.out | grep -v "100.0"

clean:
	@rm -rf bin/ build/ pkg-build/
	@echo "OK: clean up completed"

qtest:
	@#./bin/$(BINARY) -config ./assets/conf/$(BINARY).yaml -log-level debug -dry-run
	@./bin/$(BINARY) -config ./assets/conf/$(BINARY).yaml -log-level debug
