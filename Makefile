##=======================================================================##
## Makefile
## Created: Wed Aug 05 14:35:14 PDT 2015 @941 /Internet Time/
# :mode=makefile:tabSize=3:indentSize=3:
## Purpose:
##======================================================================##

SHELL=/bin/bash
PROJECT_NAME = gopass
GPATH = $(shell pwd)

.PHONY: fmt install build deps get scrape build clean

install: fmt deps
	# @GOPATH=${GPATH} go install ${PROJECT_NAME}
	@GOPATH=${GPATH} go build -o gopass main.go

build: fmt deps
	# @GOPATH=${GPATH} go build ${PROJECT_NAME}
	@GOPATH=${GPATH} go build -o gopass main.go

fmt:
	# @GOPATH=${GPATH} gofmt -s -w ${PROJECT_NAME}
	# gofmt -s -w lib/
	# gofmt -s -w main.go

deps:
	# apt-get install libgtk2.0-dev libglib2.0-dev libgtksourceview2.0-dev
	@GOPATH=${GPATH} go get -u github.com/mattn/go-gtk/gdkpixbuf
	@GOPATH=${GPATH} go get -u github.com/mattn/go-pointer
	@GOPATH=${GPATH} go get -u github.com/boltdb/bolt
	# @GOPATH=${GPATH} go get -u github.com/mattn/go-gtk

get:
	@GOPATH=${GPATH} go get ${OPTS} ${ARGS}

scrape:
	@find src -type d -name '.hg' -or -type d -name '.git' | xargs rm -rf

clean:
	@GOPATH=${GPATH} go clean
