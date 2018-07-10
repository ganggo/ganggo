# use bash instead of shell we'll
# need the ability to create arrays
SHELL=/bin/bash

version := $(shell echo -n $(VERSION) |cut -d- -f1)
package := github.com/ganggo/ganggo
srcdir := $(GOPATH)/src/$(package)

go := $(shell command -v go 2> /dev/null)
godep := $(shell command -v dep 2> /dev/null)
gobin := $(shell command -v go-bindata 2> /dev/null)
npm := $(shell command -v npm 2> /dev/null)
revel := $(shell command -v revel 2> /dev/null)
train := $(shell command -v train 2> /dev/null)

ifndef go
	$(error "go is not available: https://golang.org/dl/")
endif

define version_info
	Please define a version you want to build e.g. VERSION=v0 make
endef

define install-tools
	# go dependencies
	go get -u github.com/golang/dep/cmd/dep
	# web framework
	go get -u github.com/revel/cmd/revel
	# asset compilation
	go get -u github.com/shaoshing/train/cmd
	go build -o $$GOPATH/bin/train github.com/shaoshing/train/cmd
	# embedding binary data e.g. assets
	go get -u github.com/jteeuwen/go-bindata/...
endef

install: clean install-deps

release: precompile compile u2d-wrapper

install-deps:
ifndef npm
	$(error "npm is not available please install it first!")
endif
ifndef godep
	$(install-tools)
endif
	# Install CSS and Javascript dependencies
	cd $(srcdir) && npm install --prefix .
	# ganggo dependencies
	cd $(srcdir) && dep ensure

clean:
	rm -r tmp vendor node_modules \
		test-results bindata.go \
		updater.*.bin *.tar.gz || true

precompile:
ifndef train
	$(install-tools)
endif
	cd $(srcdir) && train -out public -source app/assets
	# Append vendor files to manifest
	sed -n '/^.*:node_modules.*/p' .include_vendor \
		| while read line; do \
			obj=($${line/:/ }); \
			dir=$(srcdir)/public/assets/vendor/$${obj[0]}; \
			mkdir -p $$dir && echo $$dir; \
			cp -v $${obj[1]} $$dir; \
		done

compile:
ifndef version
	$(error $(version_info))
endif
ifndef revel
	$(install-tools)
endif
	cp $(srcdir)/conf/app.conf.example $(srcdir)/conf/app.conf
	cd $(srcdir) && \
		APP_VERSION=$(version) revel package $(package)

test:
ifndef revel
	$(install-tools)
endif
	cd $(srcdir) && \
		go tool vet -v -all app/
	cp $(srcdir)/conf/app.conf.example $(srcdir)/conf/app.conf
	# XXX revel will not print error stacks to console
	# (see https://github.com/revel/cmd/issues/121)
	revel test $(package) ci || { \
		cd $(srcdir) && bash ci/scripts/test_results.sh ;\
		exit 1 ;\
	}

u2d-wrapper:
ifndef version
	$(error $(version_info))
endif
ifndef gobin
	$(install-tools)
endif
	mkdir -p $(srcdir)/tmp && cd $$_ && { \
		tar -x -f "../ganggo.tar.gz" ;\
		[ -f "ganggo.exe" ] && mv ganggo.exe ganggo ;\
		go-bindata -ignore \
			'src/$(package)/[vendor|node_modules|bindata.go|.*\.tar\.gz|tmp]' \
			-o ../bindata.go ganggo src/... ;\
	}
	cd $(srcdir) && go build \
		-ldflags "-X main.version=$(version)" \
		-o updater.$$GOOS-$$GOARCH$$GOARM.bin updater.go bindata.go
