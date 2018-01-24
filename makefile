# Documentation: https://www.gnu.org/software/make/manual/html_node/index.html
VERSION=0.2.0.`git rev-parse --short HEAD`
GOFLAGS=-i -v -ldflags "-s -w -X main.Version=${VERSION}" -installsuffix cgo

prowler.1m.cgo: prowler2.go
	go build ${GOFLAGS} -o prowler.1m.cgo prowler2.go

clean:
	@if [ -d release ] ; then rm -r release ; fi
	@if [ -f prowler.1m.cgo ] ; then rm prowler.1m.cgo ; fi

release: prowler.1m.cgo release/BitBarDistro.app release/bundler.sh
	@if [ -d release/Prowler.app ] ; then rm -r release/Prowler.app ; fi
	cp -R release/BitBarDistro.app release/Prowler.app
	./release/bundler.sh release/Prowler.app prowler.1m.cgo

# These files are distributed from BitBar and should not be modified
release/BitBarDistro.app:
	mkdir -p release
	wget https://github.com/matryer/bitbar/releases/download/v1.9.2/BitBarDistro-v1.9.2.zip -O release/BitBar.zip
	unzip release/BitBar.zip -d release
	rm release/BitBar.zip
release/bundler.sh:
	mkdir -p release
	wget -O release/bundler.sh https://raw.githubusercontent.com/matryer/bitbar/v1.9.2/Scripts/bitbar-bundler
	chmod u+x release/bundler.sh

.PHONY: clean release
