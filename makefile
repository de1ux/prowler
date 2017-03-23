# Documentation: https://www.gnu.org/software/make/manual/html_node/index.html
VERSION=0.0.0.`git rev-parse --short HEAD`
GOFLAGS=-i -v -ldflags "-s -w -X main.Version=${VERSION}" -installsuffix cgo

all: prowler.1m.cgo
clean:
	@if [ -f prowler.1m.cgo ] ; then rm prowler.1m.cgo ; fi

prowler.1m.cgo: prowler.go
	go build ${GOFLAGS} -o prowler.1m.cgo prowler.go

.PHONY: clean all
