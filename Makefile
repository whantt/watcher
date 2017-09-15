all: fmt lint vet watcher

LDFLAGS += -X "github.com/dearcode/watcher/util.BuildTime=$(shell date -R)"
LDFLAGS += -X "github.com/dearcode/watcher/util.BuildVersion=$(shell git rev-parse HEAD)"

VENDOR := vendor

FILES := $$(find . -name '*.go' | grep -vE 'vendor') 
SOURCE_PATH := alertor  config  editor harvester  meta  processor

golint:
	go get github.com/golang/lint/golint  

dep:
	go get -u github.com/golang/dep/cmd/dep


lint: golint
	@for path in $(SOURCE_PATH); do echo "golint $$path"; golint $$path; done;

clean:
	@rm -rf bin

fmt: 
	@for path in $(SOURCE_PATH); do echo "gofmt -s -l -w $$path";  gofmt -s -l -w $$path;  done;

vet:
	go tool vet $(FILES) 2>&1
	go tool vet --shadow $(FILES) 2>&1


watcher:
	go build -o bin/$@ -ldflags '$(LDFLAGS)' main.go 

test:
	@for path in $(SOURCE_PATH); do echo "go test ./$$path"; go test "./"$$path; done;


