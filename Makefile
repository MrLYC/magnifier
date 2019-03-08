VERSION = 0.0.1

ROOTDIR = $(shell pwd)
APPNAME = magnifier
APPPATH = github.com/mrlyc/magnifier
GODIR = /tmp/gopath
SRCDIR = ${GODIR}/src/${APPPATH}
OUTDIR = ${SRCDIR}/bin
TARGET = ${OUTDIR}/${APPNAME}

GOENV = GOPATH=${GODIR}:${GOPATH} GO15VENDOREXPERIMENT=1

GO = ${GOENV} go
DEP = ${GOENV} dep

LDFLAGS = -X ${APPPATH}/config.Version=${VERSION} -X ${APPPATH}/config.AppName=${APPNAME}
DEBUGLDFLAGS = ${LDFLAGS} -X ${APPPATH}/config.Mode=debug
RELEASELDFLAGS = -w ${LDFLAGS} -X ${APPPATH}/config.Mode=release

PACKAGES = $(shell go list ./...)

.PHONY: release
release: ${SRCDIR}
	${GO} build -i -ldflags="${RELEASELDFLAGS} -X ${APPPATH}/config.BuildHash=`git rev-parse HEAD`" -o ${TARGET} ${APPPATH}

.PHONY: build
build: ${SRCDIR}
	${GO} build -i -ldflags="${DEBUGLDFLAGS}" -o ${TARGET} ${APPPATH}

${SRCDIR}:
	mkdir -p bin
	mkdir -p `dirname "${SRCDIR}"`
	ln -s ${ROOTDIR} ${SRCDIR}

.PHONY: init
init: ${SRCDIR} update

.PHONY: update
update: ${SRCDIR}
	${DEP} ensure -v

.PHONY: test
test: ${SRCDIR}
	$(eval packages ?= ${PACKAGES})
	${GOENV} go test ${packages}

.PHONY: go-env
go-env:
	@${GOENV} go env

.PHONY: bind-builtins
bind-builtins:
	go-bindata -o ${ROOTDIR}/binding/data.go -pkg binding data